package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var simulate = flag.String("simulation", "false", "runs dataprocessor in simulation mode")

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	flag.Parse()
	var reader = io.Reader(os.Stdin)
	var samples = 40000.0 // total samples per batch to consumer before transforming

	wg := sync.WaitGroup{}
	if *simulate != "false" {
		pipeReader, pipeWriter := io.Pipe()
		wg.Add(1)
		go simulateMode(ctx, pipeWriter, samples)
		reader = pipeReader
	}

	wg.Add(1)

	cfg := getConfig()
	cfg.Samples = samples    // lets override our config to handle simulation level batche sample sizes
	cfg.BatchDuration = "5m" // lets assume we can handle 40000 sample sizes in 5m
	processor(ctx, cfg, reader)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-signals:
		cancel()
		wg.Done()
		fmt.Println(sig.String())
		os.Exit(1)
	default:
	}
}

// simulateMode starts processor in simulation mode -
func simulateMode(ctx context.Context, writer io.Writer, sampleSizes float64) {
	// wait for processor to start
	time.Sleep(1 * time.Second)
	var temps []string
	var hums []string
Loop:
	for {
		// create events to stream
		temps = generateTemps(70.0, 10.0, 160000, 4)  // we create a total of 160000/4 (amount/sensor) samples per batch(sensor) here - we need 40000 per sensor
		hums = generateHumidity(45.0, 2.0, 160000, 4) // this again uses the ^ above logic but for a new set of 4 humidity sensors
		for i, _ := range temps {
			_, _ = writer.Write([]byte(temps[i]))
			_, _ = writer.Write([]byte(hums[i]))
			time.Sleep(1 * time.Microsecond) // send in 2 logs every microsecond
		}
		break Loop // leave this for loop - do not continue making data for now
	}
}

func processor(ctx context.Context, cfg *config, reader io.Reader) error {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, reader)
	p.ProcessLogs(ctx)
	return nil
}
