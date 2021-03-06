package main

import (
	"context"
	"data-processing/batch"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

/*
func TestProcessing(t *testing.T) {
	/*
		msgs := []string{
			"reference 70.0 45.0\n",
			// "thermometer temp-1\n",
			// "2007-04-05T22:00 temp-1 72.4\n",
		}

	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = 360000

	// file := os.NewFile(uintptr(syscall.Stdin), "/dev/stdin")
	// rw := newTestReadWriter()
	// f, err := os.OpenFile("testing", os.O_CREATE|os.O_RDWR|os.O_SYNC|os.O_TRUNC|os.O_APPEND, 0664)
	f, err := os.OpenFile("testing", os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	var (
		r = bufio.NewReader(f)
		// w = bufio.NewWriter(f)
	)

	p := NewProcessor(context.Background(), cfg.Samples, time.Duration(cfg.BatchDuration), r)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go p.ProcessLogs(context.Background())
	time.Sleep(300 * time.Microsecond)

	/*
		writer := bufio.NewWriterSize(f, 500000)
		for _, v := range msgs {
			in, err := writer.WriteString(v) // this does not work
			fmt.Println(writer.Available())
			fmt.Println(in, err)
			writer.Flush()
		}

	/*
		time.Sleep(1 * time.Second)
		// NOTE: WriteFile works - but only with one operation
		os.WriteFile("testing", []byte("reference 70.0 45.0\n"), 0664) // <- this is necessary
		time.Sleep(1 * time.Second)
		os.WriteFile("testing", []byte("thermometer temp-1\n"), 0644)
		time.Sleep(2 * time.Second)

	os.WriteFile("testing", []byte("reference 70.0 45.0\n"), 0664) // <- this is necessary
	os.WriteFile("testing", []byte("reference 70.0 45.0\n"), 0664) // <- this is necessary
	os.WriteFile("testing", []byte("thermometer temp-3\n"), 0644)

	/*
		os.WriteFile("testing", []byte("reference 70.0 45.0\n"), 0664) // <- this is necessary
		os.WriteFile("testing", []byte("thermometer temp-2\n"), 0644)
		os.WriteFile("testing", []byte("thermometer temp-3\n"), 0644)
	// os.WriteFile("testing", []byte("thermometer temp-4\n"), 0644)
	// os.WriteFile("testing", []byte("thermometer temp-5\n"), 0644)
	// os.WriteFile("testing", []byte("thermometer temp-6\n"), 0644)
	// os.WriteFile("testing", []byte("thermometer temp-3\n"), 0644)
}
*/

func TestAddSensor(t *testing.T) { // DOESNT WORK
	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = "360000s"

	sensors := []string{"hum-1", "hum-2", "temp-1", "temp-2"}
	p := NewProcessor(context.Background(), cfg.Samples, cfg.BatchDuration, nil)
	for _, v := range sensors {
		go p.addSensor(context.Background(), "temp", v)
	}

	time.Sleep(2 * time.Second)

	var ok bool
	var b *batch.Batch
	for _, v := range sensors {
		b, ok = p.sensors[v]
		fmt.Println(b)
	}

	time.Sleep(2 * time.Second)

	if !ok {
		for k, _ := range p.sensors {
			spew.Sdump(k)
		}
		t.Fatalf("Failed to add all the sensors")
	}

	testBatchSensorConsumingChannels(p, t)
}

func testBatchSensorConsumingChannels(p *Processor, t *testing.T) {
	time.Sleep(2 * time.Second)
	for k, _ := range p.sensors {
		fmt.Println(k)
		p.sensors[k].Consume(5.0)
	}
}

type stringTest struct {
	whole string
	array []string
}

func newStringTest(s string) stringTest {
	sa := strings.Split(s, " ")
	return stringTest{whole: s, array: sa}
}

func TestDispatchTemperatureParse(t *testing.T) {
	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = "360000s"
	p := NewProcessor(context.Background(), cfg.Samples, cfg.BatchDuration, nil)
	s := strings.Split("thermometer temp-1", " ")
	err := p.dispatchTemperature(context.Background(), s)
	if err != nil {
		t.Fatalf("Failed to parse thermometer temp-1")
	}

	s = strings.Split("2007-04-05T22:00 temp-1 72.4", " ") // does not work
	err = p.dispatchTemperature(context.Background(), s)
	if err != nil {
		t.Fatalf("2007-04-05T22:00 temp-1 72.4")
	}
}

func TestDispatchTemperatures(t *testing.T) {
	msgs := []string{
		"thermometer temp-1",
		"2007-04-05T22:00 temp-1 72.4",
	}

	msgsTests := make([]stringTest, 0)
	for _, v := range msgs {
		msgsTests = append(msgsTests, newStringTest(v))
	}

	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = "360000s"

	p := NewProcessor(context.Background(), cfg.Samples, cfg.BatchDuration, nil)

	var err error
	for i, v := range msgsTests {
		fmt.Println("Testing", v.array)
		err = p.dispatchTemperature(context.Background(), v.array)
		time.Sleep(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed at index %v with dispatched message: %v\nBatches that exist are %v with total keys %v\nReceieved error %v", i, v.array, p.sensors, len(p.sensors), err.Error())
		}
	}

	fmt.Println(spew.Sdump(p.sensors))

	time.Sleep(2 * time.Second)
}

func TestDispatchHumidityParse(t *testing.T) {
	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = "360000s"
	p := NewProcessor(context.Background(), cfg.Samples, cfg.BatchDuration, nil)
	s := strings.Split("humidity hum-1", " ")
	err := p.dispatchHumidity(context.Background(), s)
	if err != nil {
		t.Fatalf(err.Error())
	}

	s = strings.Split("2007-04-05T22:00 hum-1 72.4", " ") // does not work
	err = p.dispatchHumidity(context.Background(), s)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestDispatchHumidities(t *testing.T) {
	msgs := []string{
		"humidity hum-1",
		"2007-04-05T22:00 hum-1 72.4",
	}

	msgsTests := make([]stringTest, 0)
	for _, v := range msgs {
		msgsTests = append(msgsTests, newStringTest(v))
	}

	cfg := new(config)
	cfg.Samples = 15
	cfg.BatchDuration = "360000s"

	p := NewProcessor(context.Background(), cfg.Samples, cfg.BatchDuration, nil)

	var err error
	for i, v := range msgsTests {
		fmt.Println("Testing", v.array)
		err = p.dispatchHumidity(context.Background(), v.array)
		if err != nil {
			t.Fatalf("Failed at index %v with dispatched message: %v\nSensors exist are %v\n", i, v.array, spew.Sdump(p.sensors))
		}
	}

	time.Sleep(2 * time.Second)
}

func TestSimulation(t *testing.T) {
	fmt.Println(generateTemps(70.0, 10.0, 100, 4))
}
