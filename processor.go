package main

import (
	"bufio"
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"data-processing/batch"
)

const (
	batchStateConsuming = "consuming"
	batchStateDone      = "done"
)

type reference struct {
	m   sync.Mutex
	set bool

	temperature float64
	humidity    float64
}

func (r reference) new(degree, humidity float64) {
	r.m.Lock()
	r.temperature = degree
	r.humidity = humidity
	r.set = true
	r.m.Unlock()
}

type Processor struct {
	m             sync.Mutex
	samples       float64
	batchDuration time.Duration
	done          chan struct{}
	cleanup       chan string

	set       bool
	reference reference

	preProcessedBuffer *list.List
	io                 io.Reader

	sensors map[string]*batch.Batch
}

func NewProcessor(ctx context.Context, samples float64, batchDuration string, reader io.Reader) *Processor {
	tp := new(Processor)
	tp.samples = samples

	duration, err := time.ParseDuration(batchDuration)
	if err != nil {
		panic(err)
	}

	tp.batchDuration = duration
	tp.set = false
	tp.done = make(chan struct{}, 1)
	tp.sensors = make(map[string]*batch.Batch)
	tp.io = reader
	tp.cleanup = make(chan string, 1)

	go tp.cleanUpSensor(ctx)

	return tp
}

func (tp *Processor) updateReference(s []string) {
	temperature, err := strconv.ParseFloat(s[1], 64)
	if err != nil {
		panic(err)
	}
	humidity, err := strconv.ParseFloat(s[2], 64)
	if err != nil {
		panic(err)
	}

	tp.reference.temperature = temperature
	tp.reference.humidity = humidity
	tp.reference.set = true
}

func (tp *Processor) checkExistingSensor(s string) bool {
	_, ok := tp.sensors[s]
	return ok
}

func (tp *Processor) sendLogs(s string, val float64) error {
	var err error
	batch, ok := tp.sensors[s]
	if ok {
		batch.Consume(val)
		return err
	}

	return errors.New("Wasn't able to send log")
}

func (tp *Processor) addSensor(ctx context.Context, s string, sensorType string) {
	ok := tp.checkExistingSensor(s)
	if !ok {
		var reference float64
		switch sensorType {
		case "temp":
			reference = tp.reference.temperature
		case "hum":
			reference = tp.reference.humidity
		}

		b := batch.NewBatch(ctx, s, sensorType, tp.samples, reference, tp.batchDuration, tp.cleanup)
		tp.m.Lock()
		tp.sensors[s] = b
		tp.m.Unlock()
		return
	}
	return
}

// TODO: Add context cancellation cleanup
func (tp *Processor) cleanUpSensor(ctx context.Context) {
	for {
		select {
		case sensor := <-tp.cleanup:
			tp.m.Lock()
			delete(tp.sensors, sensor)
			tp.m.Unlock()
		}
	}
}

func (tp *Processor) receiveFileIO(ctx context.Context) {
	r := bufio.NewReader(tp.io)
	buf := make([]byte, 20000)
	for {
		_, err := r.Read(buf)
		if err != nil && err != io.EOF {
			continue
		}

		go func() {
			fmt.Println("Processing")

			b := string(buf)
			s := strings.Split(b, " ")
			fmt.Println(s)
			if len(s) > 1 {
				fmt.Println(b)
			}
			if len(s) < 2 {
				s = append(s, " ")
			}

			_ = tp.dispatch(ctx, b, s)
		}()

		buf = make([]byte, 1024)
		continue
	}
}

func (tp *Processor) dispatch(ctx context.Context, z string, s []string) error {

	fmt.Println("DUMPING")
	var err error
	switch {
	case strings.Contains(z, "reference"):
		tp.updateReference(s)
	case strings.Contains(z, "hum"):
		err = tp.dispatchHumidity(ctx, s)
	case strings.Contains(z, " temp"):
		err = tp.dispatchTemperature(ctx, s)
	default:
		err = errors.New("Unable to dispatch")
	}
	return err
}

func (tp *Processor) dispatchTemperature(ctx context.Context, s []string) error {
	var err error
	switch {
	case s[0] == "thermometer":
		tp.addSensor(ctx, "temp", s[1])
	case strings.Contains(s[1], "temp-"):
		val, err := strconv.ParseFloat(s[2], 64)
		if err != nil {
			return err
		}
		err = tp.sendLogs(s[1], val)
	default:
		err = errors.New("wasn't able to dispatch temperature")
	}
	return err
}

func (tp *Processor) dispatchHumidity(ctx context.Context, s []string) error {
	var err error
	switch {
	case s[0] == "humidity":
		tp.addSensor(ctx, "hum", s[1])
	case strings.Contains(s[1], "hum-"):
		val, err := strconv.ParseFloat(s[2], 64)
		if err != nil {
			return err
		}
		err = tp.sendLogs(s[1], val)
	default:
		err = errors.New("wasn't able to dispatch temperature")
	}
	return err
}

func checkTemperatureSample(s []string) bool {
	if len(s) != 3 {
		return false
	}
	var ok1, ok2 bool
	_, err := strconv.ParseFloat(s[2], 64)
	if err != nil {
		ok1 = true
	}

	ok2 = strings.Contains(s[1], "temp-")
	return ok1 && ok2
}

func checkHumiditySample(s []string) bool {
	if len(s) != 3 {
		return false
	}
	var ok1, ok2 bool
	_, err := strconv.ParseFloat(s[2], 64)
	if err != nil {
		ok1 = true
	}

	ok2 = strings.Contains(s[1], "hum-")
	return ok1 && ok2
}

/*
func (tp *Processor) buildBuffer() {
	r := bufio.NewReader(os.Stdin)
	go func() {
		for !tp.reference.set {
			b, err := r.ReadString('\n')
			if err != nil {
				continue
			}
			s := strings.SplitAfter(b, " ")
			if s[0] != "reference" {
				fmt.Println("Adding to buffer")
				tp.preProcessedBuffer.PushBack(b)
				continue
			}
		}
	}()

	for {
		select {
		case <-tp.done:
			wg.Done()
			return
		}
	}
}
*/

func (tp *Processor) ProcessLogs(ctx context.Context) {
	tp.receiveFileIO(ctx)
}
