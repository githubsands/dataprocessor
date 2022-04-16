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
)

var simulate = flag.String("simulation", "false", "runs dataprocessor in simulation mode")

func main() {
	createOutputFile()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	flag.Parse()
	var reader = io.Reader(os.Stdin)
	var samples = 1000000.0 // total samples per batch to consumer before transforming
	wg := sync.WaitGroup{}
	if *simulate != "false" {
		simulator := new(simulator)
		simulator.tempSensorsAmount = 5
		pipeReader, pipeWriter := io.Pipe()
		wg.Add(1)
		go simulator.run(ctx, pipeWriter, samples)
		reader = pipeReader
	}

	cfg := getConfig()
	cfg.Samples = samples
	cfg.BatchDuration = "5m"
	wg.Add(1)
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

func processor(ctx context.Context, cfg *config, reader io.Reader) error {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, reader)
	p.ProcessLogs(ctx)
	return nil
}

func createOutputFile() {
	wd, _ := os.Getwd()
	_ = os.Remove(wd + "/output")
	_, _ = os.Create(wd + "/output")
}
