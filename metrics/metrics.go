package metrics

import (
	"time"
)

var (
	tidbQpsQuery     = ""
	tikvQpsQuery     = ""
	tidbLatencyQuery = ""
	tikvLatencyQuery = ""
)

type Metrics struct {
	start  time.Time
	end    time.Time
	source *Prometheus
}

func NewMetrics(source *Prometheus, start time.Time, end time.Time) *Metrics {
	return &Metrics{
		start,
		end,
		source,
	}
}
