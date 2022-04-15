package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`") // TODO: Add this back in
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	processor(ctx, getConfig())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-signals:
		cancel()
		time.Sleep(5 * time.Second)
		fmt.Println(sig.String())
		os.Exit(1)
	default:
	}
}

func processor(ctx context.Context, cfg *config) error {
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, os.Stdin)
	p.ProcessLogs(ctx)
	return nil
}
