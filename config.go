package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Samples       float64 `yaml:"samples"`
	BatchDuration string  `yaml:"batchduration"`
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
	yaml.Unmarshal(b, cfg)

	if cfg.Samples == 0 || cfg.BatchDuration == "" {
		panic("Must set samples and batch duration")
	}

	fmt.Println(cfg.Samples)
	fmt.Println(cfg.BatchDuration)
	return cfg
}
