package metrics

import (
	"time"
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
