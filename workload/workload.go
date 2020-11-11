package workload

import "time"

type Latency struct {
	AvgInMs float64
	P99InMs float64
}

type Record struct {
	Type    string
	Count   float64
	Latency *Latency
	Time    time.Duration
}

type Workload interface {
	Prepare() error
	Start() error
	Records() ([]*Record, error)
}
