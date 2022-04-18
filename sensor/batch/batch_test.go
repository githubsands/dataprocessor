package batch

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

func TestProcessTemperatureMath(t *testing.T) {
	b := new(Batch)
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
