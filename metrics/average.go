package metrics

import (
	"time"
)

func (m *Metrics) Average(query string) (float64, error) {
	val, err := m.source.Query(query, m.start, m.end, time.Second)
	if err != nil {
		return 0, err
	}
	values := ValuesToFloatArray(val)
	sum := 0.0
	for _, val := range values {
		sum += val.Value
	}
	return sum / float64(len(values)), nil
}
