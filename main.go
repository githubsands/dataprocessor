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

	wg := sync.WaitGroup{}
	if *simulate != "false" {
		pipeReader, pipeWriter := io.Pipe()
		wg.Add(1)
		go simulateMode(ctx, pipeWriter)
		reader = pipeReader
	}

	wg.Add(1)
	processor(ctx, getConfig(), reader)

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

func simulateMode(ctx context.Context, writer io.Writer) {
	// wait for processor to start
	time.Sleep(3 * time.Second)
	var temps []string
	var hums []string
	for {
		// start submitting data
		temps = generateTemps(70.0, 10.0, 100, 4)
		hums = generateHumidity(45.0, 2.0, 100, 4)
		for i, _ := range temps {
			_, _ = writer.Write([]byte(temps[i]))
			_, _ = writer.Write([]byte(hums[i]))
			time.Sleep(1 * time.Second)
		}
	}
}

func processor(ctx context.Context, cfg *config, reader io.Reader) error {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, reader)
	p.ProcessLogs(ctx)
	return nil
}
