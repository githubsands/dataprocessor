package main

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type config struct {
	Samples       float64 `yaml:"samples"`
	BatchDuration float64 `yaml:"batchduration"`
}

func getConfig() *config {
	cfg := new(config)
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(wd + "/config.yaml")
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(b, &cfg)
	return cfg
}

func main() {
	ctx := context.Background()
	cfg := getConfig()
	p := NewProcessor(ctx, cfg.Samples, time.Duration(cfg.BatchDuration))
	wg := sync.WaitGroup{}
	wg.Add(1)
	p.ProcessLogs(ctx)
	wg.Wait()
}
