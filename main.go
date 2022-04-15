package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`") // TODO: Add this back in
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var simulate = flag.String("simulation", "false", "runs dataprocessor in simulation mode")

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	flag.Parse()

	wg := sync.WaitGroup{}
	if *simulate != "false" {
		wg.Add(1)
		go simulateMode(ctx)
	}

	wg.Add(1)
	processor(ctx, getConfig())

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

func simulateMode(ctx context.Context) {
	writer := bufio.NewWriter(os.Stdin)
	for {
		time.Sleep(2 * time.Second)
		/*
			temps := generateTemps(70.0, 10.0, 100, 4)
			for _, v := range temps {
				_, _ = writer.WriteString(v)
			}
		*/
		_, _ = writer.WriteString("test")
	}
}

func processor(ctx context.Context, cfg *config) error {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, os.Stdin)
	p.ProcessLogs(ctx)
	return nil
}
