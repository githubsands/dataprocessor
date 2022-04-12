package main

import (
	"container/ring"
	"context"
	"fmt"
	"sync"
	"time"

	gonum "gonum.org/v1/gonum/stat"
)

const (
	temperature = "temp"
	humidity    = "hum"
)

type batch struct {
	m     sync.Mutex
	state string

	name   string
	sensor string

	tick time.Ticker

	batch *ring.Ring

	consumer chan string
	producer chan string

	cancel func()
}

func newBatch(ctx context.Context, samples float64, length time.Duration, s string) *batch {
	b := new(batch)
	b.tick = *time.NewTicker(length)
	b.batch = ring.New(int(samples))
	b.consumer = make(chan string, int(samples))
	b.producer = make(chan string, 1)
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	go b.consume(ctx)
	return b
}

func (b *batch) consume(ctx context.Context) {
	defer close(b.consumer)
	b.state = "consuming"
	for {
		select {
		// TODO
		case tempReading := <-b.consumer:
			if b.state == "consuming" {
				b.m.Lock()

				b.batch.Value = tempReading
				b.batch.Next()
				b.m.Unlock()
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

func (b *batch) process(temp float64, reference float64) {
	b.m.Lock()
	var vals []float64
	b.state = "processing"
	for i := 0; i < b.batch.Len(); i++ {
		val := b.batch.Next()
		vals = append(vals, val.Value.(float64))
	}

	var output string
	switch b.sensor {
	case temperature:
		output = b.processTemperature(vals, reference)
	case humidity:
		output = b.processHumditity(vals, reference)
	}

	b.produce(output)
	b.state = "done"
	b.m.Unlock()
	b.cancel()
}

func (b *batch) processTemperature(temps []float64, temperatureReference float64) string {
	temperatureDifferenceLow := temperatureReference - temperatureReference*0.5
	temperatureDifferenceHigh := temperatureReference + temperatureReference*0.5
	mean, std := gonum.MeanStdDev(temps, nil)
	var precision string
	switch {
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(3.0):
		precision = "ultra precise"
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(5.0):
		precision = "very price"
	default:
		precision = "precise"
	}

	return fmt.Sprintf(b.name + precision)
}

func (b *batch) processHumditity(humds []float64, humidityReference float64) string {
	humidityDifferenceLow := humidityReference - humidityReference*0.1
	humidityDifferenceHigh := humidityReference + humidityReference*0.1
	var status string = "OK"
	for _, v := range humds {
		if humidityDifferenceLow > v || humidityDifferenceHigh < v {
			status = "discard"
			break
		}
	}

	return fmt.Sprintf(b.name + status)
}

func (b *batch) produce(s string) {
	b.producer <- s
}
