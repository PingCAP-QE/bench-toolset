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

type TaggedValue struct {
	Tag   string
	Value float64
}

func WithTag(value float64, tag string) TaggedValue {
	return TaggedValue{
		Tag:   tag,
		Value: value,
	}
}

type TaggedValueSlice []TaggedValue

func (s TaggedValueSlice) Len() int {
	return len(s)
}

func (s TaggedValueSlice) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

func (s TaggedValueSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
