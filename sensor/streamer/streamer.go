package streamer

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Streamer struct {
	consumer       chan float64
	cleanup        chan<- string
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

func NewStreamer(ctx context.Context, name string, sensor string, samples float64, reference float64, length time.Duration, cleanup chan<- string) *Streamer {
	b := new(Streamer)
	b.tick = *time.NewTicker(length)
	b.name = sensor
	b.sensor = name
	b.sensorType = sensor
	b.samples = samples
	b.reference = reference
	b.consumer = make(chan float64, 1)
	b.cleanup = cleanup
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	go b.run(ctx)
	return b
}

func (b *Streamer) Consume(val float64) {
	b.consumer <- val
}

func (b *Streamer) State() string {
	return b.state
}

func (b *Streamer) run(ctx context.Context) {
	defer close(b.consumer)
	defer b.tick.Stop()
	b.state = "processing"
	go func() {
		for {
			select {
			case log := <-b.consumer:
				val := b.processHumidity(log)
				if val == "OK" {
					continue
				} else {
					b.m.Lock()
					b.state = "done"
					b.m.Unlock()
					b.write(val)
				}
			case <-b.tick.C:
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return
	}
}

func (b *Streamer) processHumidity(humidity float64) string {
	humidityDifferenceLow := b.reference - b.reference*0.1
	humidityDifferenceHigh := b.reference + b.reference*0.1
	var status string = "OK"
	if humidityDifferenceLow > humidity || humidityDifferenceHigh < humidity {
		status = "discard"
	}
	b.m.Lock()
	b.state = "done"
	b.m.Unlock()

	return b.name + " " + status
}

func (b *Streamer) write(val string) {
	wd, _ := os.Getwd()
	f, _ := os.OpenFile(wd+"/log/output-hum", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logger := log.New(f, "", 0)
	logger.Output(2, fmt.Sprintf("Processed streamer %v with %v samples as: %v\n", b.sensor, b.samples, val))

}
