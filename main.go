package main

import (
	"context"
	"flag"
	"os"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	processor()
}

func processor() {
	ctx := context.Background()
	cfg := getConfig()
	p := NewProcessor(ctx, cfg.Samples, cfg.BatchDuration, os.Stdin)
	p.ProcessLogs(ctx)
}
