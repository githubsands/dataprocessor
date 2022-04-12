package main

import (
	"bufio"
	"container/list"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	batchStateConsuming = "consuming"
	batchStateDone      = "done"
)

type reference struct {
	m   sync.Mutex
	set bool

	degrees  int
	humidity int
}

func (r reference) new(degree, humidity string) error {
	degreeNum, err := strconv.Atoi(degree)
	if err != nil {
		return err
	}

	humidityNum, err := strconv.Atoi(humidity)
	if err != nil {
		return err
	}

	r.m.Lock()
	r.degrees = degreeNum
	r.humidity = humidityNum
	r.set = true
	r.m.Unlock()

	return nil
}

type Processor struct {
	m             sync.Mutex
	samples       float64
	batchDuration time.Duration

	set       bool
	reference reference

	preProcessedBuffer *list.List

	sensors map[string]*batch
}

func NewProcessor(ctx context.Context, samples float64, batchDuration time.Duration) *Processor {
	tp := new(Processor)
	tp.samples = samples
	tp.preProcessedBuffer = list.New()
	tp.batchDuration = batchDuration
	tp.set = false
	tp.sensors = make(map[string]*batch)

	return tp
}

func (tp *Processor) addSensor(ctx context.Context, s string) {
	tp.m.Lock()
	defer tp.m.Unlock()

	val, ok := tp.sensors[s]
	if ok {
		switch {
		case val.state == "consuming":
			val.consumer <- s
			return
		default:
			return
		}
	}

	b := newBatch(ctx, tp.samples, tp.batchDuration, s)
	tp.sensors[s] = b
	return
}

func (tp *Processor) ProcessLogs(ctx context.Context) {
	r := bufio.NewReader(os.Stdin)
	wg := sync.WaitGroup{}
	wg.Add(2)

	// TODO: make its own process
	// boot the temperatureprocessor
	go func() {
		for !tp.reference.set {
			b, err := r.ReadString('\n')
			if err != nil {
				continue
			}

			ok := strings.Contains(b, "reference")
			if ok {
				continue
			}

			s := strings.SplitAfter(b, " ")
			tp.reference.new(s[1], s[2])
		}

		return
	}()

	// TODO: make its own function
	go func() {
		for !tp.reference.set {
			b, err := r.ReadString('\n')
			if err != nil {
				continue
			}
			s := strings.SplitAfter(b, " ")
			if s[0] != "reference" {
				tp.preProcessedBuffer.PushBack(b)
				continue
			}
		}

		return
	}()

	wg.Wait()

	// TODO: Make its own function
	for {
		go func() {
			for e := tp.preProcessedBuffer.Front(); e != nil; e = e.Next() {
				tp.addSensor(ctx, e.Value.(string))
			}
			return
		}()
		sensor, err := r.ReadString('\n')
		if err != nil {
			continue
		}

		tp.addSensor(ctx, sensor)
		continue
	}
}
