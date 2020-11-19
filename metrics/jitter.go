package metrics

import (
	"math"
	"time"
)

type JitterResult struct {
	Sd  float64
	Max float64
}

func (m *Metrics) Jitter(query string) (*JitterResult, error) {
	val, err := m.source.Query(query, m.start, m.end, time.Second)
	if err != nil {
		return nil, err
	}
	values := ValuesToFloatArray(val)
	return m.CalculateJitter(values), nil
}

func (m *Metrics) CalculateJitter(values []float64) *JitterResult {
	sum := float64(0)
	count := float64(len(values))
	for _, v := range values {
		sum += v
	}
	avg := sum / count
	powDeltaSum := float64(0)
	jitterMax := float64(0)
	for _, v := range values {
		jitter := math.Abs(v - avg)
		if jitter > jitterMax {
			jitterMax = jitter
		}
		powDeltaSum += math.Pow(jitter, 2)
	}
	return &JitterResult{
		Sd:  math.Sqrt(powDeltaSum / count) / avg,
		Max: jitterMax,
	}
}
