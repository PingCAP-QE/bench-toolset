package bench

import (
	"strconv"
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

func (b *GcBench) Prepare() error {
	return b.load.Prepare()
}

func (b *GcBench) Run() error {
	b.start = time.Now()
	err := b.load.Start()
	if err != nil {
		return err
	}
	b.end = time.Now()
	return nil
}

func (b *GcBench) Results() ([]*Result, error) {
	records, err := b.load.Records()
	if err != nil {
		return nil, err
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

	return []*Result{
		{"tps-jitter-sd", strconv.FormatFloat(countJitter.Sd, 'f', 2, 64)},
		{"tps-jitter-max", strconv.FormatFloat(countJitter.Max, 'f', 2, 64)},

		{"avg-lat-jitter-sd", strconv.FormatFloat(avgLatJitter.Sd, 'f', 2, 64)},
		{"avg-lat-jitter-max", strconv.FormatFloat(avgLatJitter.Max, 'f', 2, 64)},

		{"p99-lat-jitter-sd", strconv.FormatFloat(p99LatJitter.Sd, 'f', 2, 64)},
		{"p99-lat-jitter-max", strconv.FormatFloat(p99LatJitter.Max, 'f', 2, 64)},
	}, nil
}
