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
	set bool

	temperature float64
	humidity    float64
}

func (r reference) new(degree, humidity float64) {
	r.temperature = degree
	r.humidity = humidity
	r.set = true
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
	tp.cleanup = make(chan string, 10000)

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

func (tp *Processor) sendLogs(sensorName string, val float64) error {
	var err error
	fmt.Println(sensorName)
	batch, ok := tp.sensors[sensorName]
	if ok {
		batch.Consume(val)
		return err
	}

	return errors.New(fmt.Sprintf("%v does not exist. Unable to send batch data\n", sensorName))
}

func (tp *Processor) addSensor(ctx context.Context, sensorName string, sensorType string) error {
	ok := tp.checkExistingSensor(sensorName)
	if !ok {
		var reference float64
		switch sensorType {
		case "temp":
			reference = tp.reference.temperature
		case "hum":
			reference = tp.reference.humidity
		default:
			break
		}

		b := batch.NewBatch(ctx, sensorName, sensorType, tp.samples, reference, tp.batchDuration, tp.cleanup)
		tp.m.Lock()
		fmt.Println("Adding sensor", sensorName)
		tp.sensors[sensorName] = b
		tp.m.Unlock()
		return nil
	}

	return errors.New(fmt.Sprintf("Unable to send data - sensorType, %v is not implemented", sensorType))
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
	buf := make([]byte, 0, 30) // log has a max size of 16 - buffer 400 of these
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if err != nil && err != io.EOF {
			continue
		}

		s1, s2, err := tp.processString(buf[:])
		if err != nil {
			continue
		}

		err = tp.dispatch(ctx, s1, s2)
		if err != nil {
			fmt.Printf("Wasn't able to dispatch %v: %v", s1, err.Error())
			continue
		}
	}
}

func (tp *Processor) processString(buf []byte) (string, []string, error) {
	s1 := string(buf)

	s2 := strings.Split(s1, " ")
	if len(s2) < 2 {
		s2 = append(s2, " ")
	}

	// TODO: This doesn't work as intended
	if s1 == "" {
		return "", nil, errors.New(fmt.Sprintf("Bad string input given"))
	}
	return s1, s2, nil
}

// TODO: This function should be disolved with multiplexed IO
func (tp *Processor) dispatch(ctx context.Context, z string, s []string) error {
	var err error
	switch {
	case strings.Contains(z, "reference"):
		tp.updateReference(s)
	case strings.Contains(z, "hum"):
		err = tp.dispatchHumidity(ctx, s)
	case strings.Contains(z, " temp"):
		err = tp.dispatchTemperature(ctx, s)
	default:
		err = errors.New(fmt.Sprintf("Given string %v cannot be dispatched\n", z))
	}

	return err
}

func (tp *Processor) dispatchTemperature(ctx context.Context, s []string) error {
	var err error
	switch {
	case s[0] == "thermometer":
		tp.addSensor(ctx, s[1], "temp")
	case strings.Contains(s[1], "temp-"):
		fmt.Println("ADDING TEMP READ", s[2])
		s[2] = strings.TrimRight(s[2], "\n")
		val, err := strconv.ParseFloat(s[2], 64)
		if err != nil {
			fmt.Println("Got", val, err)
			return err
		}
		fmt.Println("Got", val)
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
		tp.addSensor(ctx, s[1], "hum")
	case strings.Contains(s[1], "hum-"):
		s[2] = strings.TrimRight(s[2], "\n")
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

func (tp *Processor) ProcessLogs(ctx context.Context) {
	tp.receiveFileIO(ctx)
}
