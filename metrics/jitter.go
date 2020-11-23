package metrics

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

type JitterResult struct {
	Sd     float64
	Max    float64
	KMean  float64
	Random float64
}

func (m *Metrics) Jitter(query string) (*JitterResult, error) {
	val, err := m.source.Query(query, m.start, m.end, time.Second)
	fmt.Println(val)
	if err != nil {
		return nil, err
	}
	values := ValuesToFloatArray(val)
	fmt.Println(values)
	results, _ := CalculateJitter(values, 0, 0)
	return results, nil
}

func CalculateJitter(values []float64, kNumber int, percent float64) (*JitterResult, float64) {
	sort.Float64s(values)
	sd := stdev(values)
	avg := avgf(values)
	jMax := (values[len(values)-1] - values[0]) / avg
	var kMean float64
	if kNumber != 0 && len(values) > kNumber {
		kMin := avgf(values[:kNumber])
		kMax := avgf(values[len(values)-kNumber:])
		kMean = (kMax - kMin) / avg
	}
	var randSum float64
	if percent > 0 && percent < 100 {
		rand.Seed(time.Now().UnixNano())
		count := percent / 100 * float64(len(values))
		for i := 0; i < int(count); i++ {
			r := rand.Intn(len(values))
			randSum += math.Abs(values[r] - avg)
			values[len(values)-1], values[r] = values[r], values[len(values)-1]
			values = values[:len(values)-1]
		}
		randSum /= count
	}
	return &JitterResult{
		Sd:    sd / avg,
		Max:   jMax,
		KMean: kMean,
		Random: randSum / avg,
	}, avg
}

func (m *Metrics) TiDBCollectJitter(intervalSecs uint64) *JitterResult {
	return nil
}

func stdev(values []float64) float64 {
	count := float64(len(values))
	avg := avgf(values)
	powDeltaSum := float64(0)
	for _, v := range values {
		jitter := math.Abs(v - avg)
		powDeltaSum += math.Pow(jitter, 2)
	}
	return math.Sqrt(powDeltaSum / (count - 1))
}

func avgf(values []float64) float64 {
	sum := float64(0)
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
