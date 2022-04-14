package main

import (
	"bufio"
	"container/list"
	"context"
	"fmt"
	"io"
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
	done          chan struct{}
	cleanup       chan string

	set       bool
	reference reference

	preProcessedBuffer *list.List
	file               io.Reader

	sensors map[string]*batch
}

func NewProcessor(ctx context.Context, samples float64, batchDuration time.Duration, file io.Reader) *Processor {
	tp := new(Processor)
	tp.samples = samples
	tp.preProcessedBuffer = list.New()
	tp.batchDuration = batchDuration
	tp.set = false
	tp.done = make(chan struct{}, 1)
	tp.sensors = make(map[string]*batch)
	tp.file = file
	tp.cleanup = make(chan string, 1)

	go tp.cleanUpSensor(ctx)

	return tp
}

func (tp *Processor) checkExistingSensor(s string) bool {
	_, ok := tp.sensors[s]
	fmt.Println("Checking existing sensor")
	return ok
}

func (tp *Processor) sendLogs(s string, val float64) bool {
	var success bool = true
	batch, ok := tp.sensors[s]
	if ok {
		batch.consumer <- val
		return success
	}

	return false
}

func (tp *Processor) addSensor(ctx context.Context, s string, sensorType string) {
	ok := tp.checkExistingSensor(s)
	if !ok {
		b := newBatch(ctx, s, sensorType, tp.samples, float64(tp.reference.humidity), tp.batchDuration, tp.cleanup)
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
	r := bufio.NewReader(tp.file)
	buf := make([]byte, 20000)
	for {
		_, err := r.Read(buf)
		if err != nil && err != io.EOF {
			continue
		}

		b := string(buf)
		s := strings.Split(b, " ")
		if len(s) > 1 {
			fmt.Println(b)
		}
		if len(s) < 2 {
			s = append(s, " ")
		}

		_ = tp.dispatch(ctx, b, s)
		buf = make([]byte, 1024)
		continue
	}

	fmt.Println("Closing reference receiver")
	return
}

func (tp *Processor) dispatch(ctx context.Context, z string, s []string) bool {
	var dispatched = true
	switch {
	case strings.Contains(z, "reference"):
		tp.reference.new(s[1], s[2])
		tp.reference.set = true
	case strings.Contains(z, "humidity"):
		dispatched = tp.dispatchHumidity(ctx, s)
	case strings.Contains(z, "thermometer"):
		dispatched = tp.dispatchTemperature(ctx, s)
	default:
		dispatched = false
	}
	return dispatched
}

func (tp *Processor) dispatchTemperature(ctx context.Context, s []string) bool {
	success := true
	switch {
	case s[0] == "thermometer":
		go tp.addSensor(ctx, "temp", s[1])
	case checkTemperatureSample(s):
		val, _ := strconv.ParseFloat(s[2], 64)
		success = tp.sendLogs(s[1], val)
	default:
		success = false
	}
	return success
}

func (tp *Processor) dispatchHumidity(ctx context.Context, s []string) bool {
	success := true
	switch {
	case s[0] == "humidity":
		go tp.addSensor(ctx, "hum", s[1])
	case checkHumiditySample(s):
		val, _ := strconv.ParseFloat(s[2], 64)
		success = tp.sendLogs(s[1], val)
	default:
		success = false
	}
	return success
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
