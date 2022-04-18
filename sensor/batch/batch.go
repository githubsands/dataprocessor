package batch

import (
	"container/ring"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	gonum "gonum.org/v1/gonum/stat" // TODO: May be better to add floats as they are received...
)

const (
	temperature = "temp"
)

type Batch struct {
	consumer       chan float64
	Batch          *ring.Ring
	cleanup        chan<- string
	pool           sync.Pool
	producer       chan string
	cancel         func()
	state          string
	sensor         string
	sensorType     string
	name           string
	tick           time.Ticker
	reference      float64
	samples        float64
	currentSamples float64
	m              sync.Mutex
}

func NewBatch(ctx context.Context, name string, sensor string, samples float64, reference float64, length time.Duration, cleanup chan<- string) *Batch {
	b := new(Batch)
	b.tick = *time.NewTicker(length)
	b.name = sensor
	b.sensor = name
	b.sensorType = sensor
	b.Batch = ring.New(int(samples))
	b.currentSamples = 0
	b.samples = samples
	b.reference = reference
	b.consumer = make(chan float64, int(samples))
	b.producer = make(chan string, 1)
	b.cleanup = cleanup
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	go b.run(ctx)
	go b.write(ctx)
	return b
}

func (b *Batch) State() string {
	return b.state
}

func (b *Batch) Consume(val float64) {
	b.consumer <- val
}

func (b *Batch) run(ctx context.Context) {
	defer close(b.consumer)
	defer b.tick.Stop()
	b.state = "processing"
	go func() {
		for {
			select {
			case log := <-b.consumer:
				if b.currentSamples == b.samples {
					b.state = "processing"
					b.process(b.reference)
					return
				}
				b.Batch.Value = log
				b.Batch.Next()
				b.m.Lock()
				b.currentSamples++
				b.m.Unlock()
			case <-b.tick.C:
				b.process(b.reference)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return
	}
}

func (b *Batch) process(reference float64) {
	b.m.Lock()
	var vals []float64
	b.state = "processing"
	for i := 0; i < b.Batch.Len(); i++ {
		val := b.Batch
		vals = append(vals, val.Value.(float64)) // TODO: I assume this uses reflection so not very optimal
		_ = b.Batch.Prev()
	}

	var output string
	switch b.sensorType {
	case temperature:
		output = b.processTemperature(vals, reference)
	default:
		return
	}

	b.produce(output)

	b.m.Lock()
	b.state = "done"
	b.m.Unlock()

	b.cleanup <- b.name
	b.cancel()
	return
}

func (b *Batch) processTemperature(temps []float64, temperatureReference float64) string {
	temperatureDifferenceLow := temperatureReference - temperatureReference*0.5
	temperatureDifferenceHigh := temperatureReference + temperatureReference*0.5
	mean, std := gonum.MeanStdDev(temps, nil)
	var precision string
	switch {
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(3.0):
		precision = "ultra precise"
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(5.0):
		precision = "very precise"
	default:
		precision = "precise"
	}

	return b.name + " " + precision
}

func (b *Batch) produce(s string) {
	b.producer <- s
}

func (b *Batch) write(ctx context.Context) {
	defer close(b.producer)
	for {
		select {
		case val := <-b.producer:
			wd, _ := os.Getwd()
			f, _ := os.OpenFile(wd+"/log/output-temp", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			logger := log.New(f, "", 0)
			logger.Output(2, fmt.Sprintf("Processed batch %v with %v samples as: %v\n", b.sensor, b.samples, val))
		case <-ctx.Done():
			return
		}
	}
}
