package main

import (
	"fmt"
	"strings"

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
	temperatureLogSensor   = "thermometer temp-N"
	temperatureLogSkeleton = "2007-04-05T22:00 temp-N TEMP"
)

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
	fmt.Println(inputTypeString)
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
		fmt.Println(temperatureLogSkeleton)
		temps = append(temps, generateInput(temperatureLogSkeleton, v))
	}

	return temps
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
