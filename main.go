package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var simulate = flag.String("simulation", "false", "runs dataprocessor in simulation mode")
var sensors = flag.Int("sensors", 1, "creates N humidity and N temperature sensors")
var sampleSize = flag.Float64("samples", 5.0, "creates N samples for humidity and temperature")

func main() {
	resetOutputFiles()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	flag.Parse()

	var reader = io.Reader(os.Stdin)
	var samples = *sampleSize // total samples per batch to consumer before transforming
	var tempSensors = *sensors
	var humiditySamples = *sensors

	wg := sync.WaitGroup{}
	if *simulate != "false" {
		wg.Add(1)
		go func() {
			_, pipeReader := newSimulator(ctx, samples, tempSensors, humiditySamples)
			reader = pipeReader
		}()
	}

	cfg := getConfig()
	cfg.Samples = samples
	cfg.BatchDuration = "5m"
	wg.Add(2)
	go processor(ctx, cfg, reader)

	wg.Wait()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case _ = <-signals:
		cancel()
		wg.Done()
		os.Exit(1)
	default:
	}
}

func processor(ctx context.Context, cfg *config, reader io.Reader) {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, reader)
	p.ProcessLogs(ctx)
}

func resetOutputFiles() {
	wd, _ := os.Getwd()
	_ = os.Remove(wd + "/log/output-hum")
	_ = os.Remove(wd + "/log/output-temp")
}
