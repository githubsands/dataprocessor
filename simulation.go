package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"gonum.org/v1/gonum/stat/distuv"
)

type replace struct {
	old      string
	replaced string
}

var (
	referenceSkeleton       = "reference TEMP HUM"
	referenceReplacementSet = map[string]string{
		"TEMP": "",
		"HUM":  "",
	}
)

var (
	humiditySkeletonSensor = "humidity hum-N"
	humiditySkeletonLog    = "2007-04-05T22:00 hum-N HUM"
)

var (
	temperatureLogSensorSkeleton = "thermometer temp-N"
	temperatureLogSkeleton       = "2007-04-05T22:00 temp-N TEMP"
)

type simulator struct {
	sampleSizes float64

	tempSensorsAmount     int
	humiditySensorsAmount int
}

func newSimulator(ctx context.Context, sampleSizes float64, tempSensorsAmount int, humiditySensorsAmount int) (*simulator, io.Reader) {
	simulator := new(simulator)
	simulator.tempSensorsAmount = tempSensorsAmount
	simulator.humiditySensorsAmount = humiditySensorsAmount
	simulator.sampleSizes = sampleSizes
	pipeReader, pipeWriter := io.Pipe()
	go simulator.run(ctx, pipeWriter, sampleSizes)

	return simulator, pipeReader
}

// simulateMode starts processor in simulation mode -
func (s *simulator) run(ctx context.Context, writer io.Writer, sampleSizes float64) {
	time.Sleep(1 * time.Second) // wait for processor to start
	var referenceSet bool = false
	for {
		if !referenceSet {
			reference := generateReference(1)
			num, _ := writer.Write([]byte(reference))
			if num != 0 {
				referenceSet = true
			}
			tempSensors := generateTempSensors(s.tempSensorsAmount)
			for i := range tempSensors {
				writer.Write([]byte(tempSensors[i]))
			}
			humiditySensors := generateHumiditySensors(s.humiditySensorsAmount)
			for i := range humiditySensors {
				writer.Write([]byte(humiditySensors[i]))
			}
		}

		temps := generateTemps(80.0, 30.0, 500000, s.tempSensorsAmount) // 100000 samples for each temp sensor
		for i := range temps {
			writer.Write([]byte(temps[i]))
		}

		hums := generateHumidity(80.0, 30.0, 500000, s.humiditySensorsAmount) // 100000 samples for each temp sensor
		for i := range hums {
			writer.Write([]byte(hums[i]))
		}
	}
}

func NewTempeartureSensorReplacementSet() map[string]string {
	m := make(map[string]string)
	m["N"] = ""
	return m
}

func NewHumiditySensorReplacementSet() map[string]string {
	m := make(map[string]string)
	m["N"] = ""
	return m
}

func NewTempeartureReplacementSet() map[string]string {
	m := make(map[string]string)
	m["N"] = ""
	m["TEMP"] = ""
	return m
}

func NewHumidityReplacementSet() map[string]string {
	m := make(map[string]string)
	m["N"] = ""
	m["HUM"] = ""
	return m
}

func generateInput(inputTypeString string, replace map[string]string) string {
	if len(replace) != 0 {
		for k, v := range replace {

			inputTypeString = strings.Replace(inputTypeString, k, v, 1)
			delete(replace, k)
			inputTypeString = generateInput(inputTypeString, replace)
		}
	}

	return inputTypeString
}

func generateReference(amount int) string {
	referenceReplacementSet["TEMP"] = "70.0"
	referenceReplacementSet["HUM"] = "45.0"

	return generateInput(referenceSkeleton, referenceReplacementSet)
}

func generateTempSensors(amount int) []string {
	var tempSensors []string = make([]string, amount)
	var tr map[string]string
	var tempSensor string
	for i := 1; i < amount+1; i++ {
		tr = NewTempeartureSensorReplacementSet()
		tr["N"] = fmt.Sprintf("%d", i)
		tempSensor = generateInput(temperatureLogSensorSkeleton, tr)
		tempSensors = append(tempSensors, tempSensor)
	}

	return tempSensors
}

func generateTemps(temperature float64, sigma float64, sampleSize int, sensors int) []string {
	dist := distuv.Normal{
		Mu:    temperature,
		Sigma: sigma,
	}
	sensor := 1
	diff := sampleSize / sensors
	var temps []string
	var tr map[string]string
	var trs []map[string]string
	for i := 0; i < sampleSize; i++ {
		if i == diff*sensor {
			sensor++
		}

		tr = NewTempeartureReplacementSet()
		tr["N"] = fmt.Sprintf("%d", sensor)
		tr["TEMP"] = fmt.Sprintf("%.2f", dist.Rand())
		trs = append(trs, tr)
	}

	for _, v := range trs {
		temps = append(temps, generateInput(temperatureLogSkeleton, v))
	}

	return temps
}

func generateHumiditySensors(amount int) []string {
	var humiditySensors []string = make([]string, amount)
	var tr map[string]string
	var humiditySensor string
	for i := 1; i < amount+1; i++ {
		tr = NewHumiditySensorReplacementSet()
		tr["N"] = fmt.Sprintf("%d", i)
		humiditySensor = generateInput(humiditySkeletonSensor, tr)
		humiditySensors = append(humiditySensors, humiditySensor)
	}

	return humiditySensors
}

func generateHumidity(humidity float64, sigma float64, sampleSize int, sensors int) []string {
	dist := distuv.Normal{
		Mu:    humidity,
		Sigma: sigma,
	}
	sensor := 1
	diff := sampleSize / sensors
	var temps []string
	var tr map[string]string
	var trs []map[string]string
	for i := 0; i < sampleSize+1; i++ {
		if i == diff*sensor {
			sensor++
		}

		tr = NewHumidityReplacementSet()
		tr["N"] = fmt.Sprintf("%d", sensor)
		tr["HUM"] = fmt.Sprintf("%.2f", dist.Rand())
		trs = append(trs, tr)
	}

	for _, v := range trs {
		temps = append(temps, generateInput(humiditySkeletonLog, v))
	}

	return temps
}
