package metrics

import (
	"math"
	"time"
)

type AggregateResult struct {
	Avg float64
	Max float64
	Min float64
}

func (m *Metrics) Aggregate(query string) (*AggregateResult, error) {
	val, err := m.source.Query(query, m.start, m.end, time.Second)
	if err != nil {
		return nil, err
	}
	values := ValuesToFloatArray(val)
	sum := 0.0
	max := -math.MaxFloat64
	min := math.MaxFloat64
	for _, val := range values {
		sum += val.Value
		max = math.Max(max, val.Value)
		min = math.Min(min, val.Value)
	}
	return &AggregateResult{
		Avg: sum / float64(len(values)),
		Max: max,
		Min: min,
	}, nil
}
