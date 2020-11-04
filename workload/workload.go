package workload

import "time"

type Latency struct {
	Avg float64
	P99 float64
}

type Record struct {
	Type    string
	Count   float64
	Latency *Latency
	Time    time.Duration
}

type Workload interface {
	Start() error
	Records() ([]*Record, error)
}
