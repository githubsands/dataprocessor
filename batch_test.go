package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestBatchRunTemp(t *testing.T) {
	defer goleak.VerifyNone(t) // test if all goroutines are cleaned up

	var duration time.Duration = 40 * time.Second
	var sampleSize = 5.0
	var cleanup = make(chan string, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	batch := newBatch(context.TODO(), "temp-1", "temp", sampleSize, 40.0, duration, cleanup)
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0 // these inputs exceed the sampleSize so they will not make it to processing

	var actual1, actual2 string
Receive:
	for {
		select {
		case outcome := <-batch.producer:
			actual1 = outcome
			fmt.Println("Received", actual1, "from producer")
		case cleaned := <-cleanup:
			actual2 = cleaned
			fmt.Println("Received", actual2, "from producer")
			wg.Done()
			break Receive
		}
	}

	wg.Wait()

	expected1 := "temp-1 ultra precise"
	if expected1 != actual1 {
		t.Logf(actual1)
		t.Fatalf("Expected %v, got actual %v", expected1, actual1)
	}

	fmt.Println(actual2)
	expected2 := "temp-1"
	if expected2 != actual2 {
		t.Fatalf("Expected %v, got actual %v", expected2, actual2)
	}

	testBatchChannelCleanup(batch, t)
}

func testBatchChannelCleanup(batch *batch, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Recover should of not been nil due to having send on a closed channel")
		}
	}()

	batch.consumer <- 40.0 // should cause a panic but we will recover
}

/*
func TestBatchRunHumidity(t *testing.T) {
	var duration time.Duration = 40 * time.Second
	var sampleSize = 5.0
	var cleanup = make(chan string, 1)
	batch := newBatch(context.TODO(), "hum-1", "hum", sampleSize, 40.0, duration, cleanup)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go batch.run(context.TODO())
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0 // these inputs exceed the sampleSize so they will not make it to processing
	batch.consumer <- 40.0

	var actual1, actual2 string
	select {
	case actual1 := <-batch.producer:
	case actual2 := <-cleanup:
		break
	}

	expected1 := "ultra precise"
	if expected1 != actual1 {
		t.Fatalf("Expected %v, got actual %v", expected2, actual2)
	}

	expected2 := "hum-1"
	if expected2 != actual2 {
		t.Fatalf("Expected %v, got actual %v", expected2, actual2)
	}

	fmt.Println(expected1, actual1, expected2, actual2)
}
*/

func TestProcessTemperatureMath(t *testing.T) {
	b := new(batch)
	b.name = "temp-1"
	floats := []float64{70.0, 70.0, 70.0, 70.0}
	actual := b.processTemperature(floats, 70)
	expected := "temp-1 ultra precise"
	if actual != expected {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	floats = []float64{63.0, 68.0, 72.0, 73.0}
	actual = b.processTemperature(floats, 70)
	expected = "temp-1 very precise"
	if actual != expected {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	floats = []float64{60.0, 65.0, 70.0, 80.0}
	actual = b.processTemperature(floats, 70)
	expected = "temp-1 precise"
	if actual != expected {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}
}

func TestProcessHumidityMath(t *testing.T) {
	b := new(batch)
	b.name = "hum-1"
	humids := []float64{40.0, 40.0, 40.0, 40.0}
	actual := b.processHumidity(humids, 40.0)
	expected := "hum-1 OK"
	if expected != actual {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	humids = []float64{40.0, 40.0, 40.2, 40.0}
	actual = b.processHumidity(humids, 40.0)
	expected = "hum-1 OK"
	if expected != actual {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	humids = []float64{40.0, 39.8, 40.0, 40.0}
	actual = b.processHumidity(humids, 40.0)
	expected = "hum-1 OK"
	if expected != actual {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	humids = []float64{35.99, 40.0, 40.0, 40.0, 40.40}
	actual = b.processHumidity(humids, 40.0)
	expected = "hum-1 discard"
	if expected != actual {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}

	humids = []float64{40.0, 40.0, 40.0, 40.0, 44.01}
	actual = b.processHumidity(humids, 40.0)
	expected = "hum-1 discard"
	if expected != actual {
		t.Fatalf("Expected %v, actual %v", expected, actual)
	}
}
