package main

type Sensor interface {
	State() string
	Consume(float64)
}
