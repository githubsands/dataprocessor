package streamer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestStreamerRun(t *testing.T) {
	defer goleak.VerifyNone(t) // test if all goroutines are cleaned up

	var duration time.Duration = 40 * time.Second
	var sampleSize = 5.0
	var cleanup = make(chan string, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	batch := NewBatch(context.TODO(), "temp", "temp-1", sampleSize, 40.0, duration, cleanup)
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0
	batch.consumer <- 40.0 // these inputs exceed the sampleSize so they will not make it to processing

	var actual1, actual2 string
Loop:
	for {
		select {
		case outcome := <-batch.producer:
			actual1 = outcome
			fmt.Println("Received", actual1, "from producer")
		case cleaned := <-cleanup:
			actual2 = cleaned
			fmt.Println("Received", actual2, "from producer")
			wg.Done()
			break Loop
		}
	}

	t.Logf("Waiting")
	wg.Wait()
	t.Logf("Waiting")

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

func testBatchChannelCleanup(batch *Batch, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Recover should of not been nil due to having send on a closed channel")
		}
	}()

	batch.consumer <- 40.0 // should cause a panic but we will recover
}

func TestProcessHumidityMath(t *testing.T) {
	b := new(Batch)
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
