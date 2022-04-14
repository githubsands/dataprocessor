package batch

import (
	"container/ring"
	"context"
	"fmt"
	"sync"
	"time"

	gonum "gonum.org/v1/gonum/stat" // TODO: May be better to add floats as they are received...
)

const (
	temperature = "temp"
	humidity    = "hum"
)

type Batch struct {
	m     sync.Mutex
	state string

	name           string
	sensor         string
	currentSamples float64
	samples        float64
	reference      float64

	tick time.Ticker // TODO: Could just use a context timeout here

	Batch *ring.Ring // TODO: Change to just a chan or other methods noted in the README

	consumer chan float64 // ... TODO: Could just use a channel instead of ring buffer
	producer chan string

	cleanup chan<- string

	cancel func()
}

func NewBatch(ctx context.Context, name string, sensor string, samples float64, reference float64, length time.Duration, cleanup chan<- string) *Batch {
	b := new(Batch)
	b.tick = *time.NewTicker(length)
	b.name = sensor
	b.sensor = name
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
	return b
}

func (b *Batch) Consume(val float64) {
	b.consumer <- val
}

func (b *Batch) run(ctx context.Context) {
	defer close(b.consumer)
	defer b.tick.Stop()

	b.state = "consuming"
	go func() {
		for {
			select {
			case tempReading := <-b.consumer:
				if b.currentSamples == b.samples {
					b.process(b.reference)
					return
				}
				b.Batch.Value = tempReading
				b.Batch.Next()
				b.m.Lock()
				fmt.Println("FUUUCK-3", tempReading)
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
	switch b.sensor {
	case temperature:
		output = b.processTemperature(vals, reference)
		fmt.Println("PROCESSING")
	case humidity:
		output = b.processHumidity(vals, reference)
	default:
		// discard the sensor given is not implemented
		return
	}

	b.produce(output)

	b.state = "done"
	fmt.Println("cleaning up")
	b.cleanup <- b.name
	b.cancel()
	return
}

//TODO: Batch should only process - not be aware which sensor its processing. possibly take in a first class function here
func (b *Batch) processTemperature(temps []float64, temperatureReference float64) string {
	temperatureDifferenceLow := temperatureReference - temperatureReference*0.5
	temperatureDifferenceHigh := temperatureReference + temperatureReference*0.5
	mean, std := gonum.MeanStdDev(temps, nil)
	var precision string
	switch {
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(3.0):
		fmt.Println("made it --")
		precision = "ultra precise"
	case temperatureDifferenceLow <= mean && mean <= temperatureDifferenceHigh && std < float64(5.0):
		fmt.Println("made it --")
		precision = "very precise"
	default:
		fmt.Println("made it --")
		precision = "precise"
	}

	return fmt.Sprintf(b.name + " " + precision)
}

//TODO: Batch should only process - not be aware which sensor its processing. possibly take in a first class function here
func (b *Batch) processHumidity(humds []float64, humidityReference float64) string {
	humidityDifferenceLow := humidityReference - humidityReference*0.1
	humidityDifferenceHigh := humidityReference + humidityReference*0.1
	fmt.Println(humidityDifferenceHigh)
	var status string = "OK"
	for _, v := range humds {
		if humidityDifferenceLow > v || humidityDifferenceHigh < v {
			status = "discard"
			break
		}
	}

	return fmt.Sprintf(b.name + " " + status)
}

func (b *Batch) produce(s string) {
	b.producer <- s
}
