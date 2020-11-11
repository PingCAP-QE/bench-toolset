package bench

import (
	"fmt"
	"time"

	"github.com/5kbpers/stability_bench/metrics"
	"github.com/5kbpers/stability_bench/workload"
)

type GcBench struct {
	load  workload.Workload
	start time.Time
	end   time.Time
}

func NewGcBench(load workload.Workload) *GcBench {
	return &GcBench{
		load: load,
	}
}

func (b *GcBench) Run() error {
	err := b.load.Prepare()
	if err != nil {
		return err
	}
	b.start = time.Now()
	err = b.load.Start()
	if err != nil {
		return err
	}
	b.end = time.Now()
	return nil
}

func (b *GcBench) Report() (string, error) {
	records, err := b.load.Records()
	if err != nil {
		return "", err
	}
	m := metrics.NewMetrics(nil, b.start, b.end)
	counts := make([]float64, len(records))
	avgLats := make([]float64, len(records))
	p99Lats := make([]float64, len(records))

	for i, r := range records {
		counts[i] = r.Count
		avgLats[i] = r.Latency.AvgInMs
		p99Lats[i] = r.Latency.P99InMs
	}

	countJitter := m.CalculateJitter(counts)
	avgLatJitter := m.CalculateJitter(avgLats)
	p99LatJitter := m.CalculateJitter(p99Lats)

	countRes := fmt.Sprintf("tps jitter-sd %.2f, jitter-max %.2f\n", countJitter.Sd, countJitter.Max)
	avgLatRes := fmt.Sprintf("avg_lat jitter-sd %.2f, jitter-max %.2f\n", avgLatJitter.Sd, avgLatJitter.Max)
	p99LatRes := fmt.Sprintf("p99_lat jitter-sd %.2f, jitter-max %.2f\n", p99LatJitter.Sd, p99LatJitter.Max)

	return countRes + avgLatRes + p99LatRes, nil
}
