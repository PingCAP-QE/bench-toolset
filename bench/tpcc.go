package bench

import (
	"strconv"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

type TpccBench struct {
	load  workload.Workload
	start time.Time
	end   time.Time
}

func NewTpccBench(load workload.Workload) *TpccBench {
	return &TpccBench{
		load: load,
	}
}

func (b *TpccBench) Prepare() error {
	return b.load.Prepare()
}

func (b *TpccBench) Run() error {
	b.start = time.Now()
	err := b.load.Start()
	if err != nil {
		return err
	}
	b.end = time.Now()
	return nil
}

func (b *TpccBench) Results() ([]*Result, error) {
	records, err := b.load.Records()
	if err != nil {
		return nil, err
	}
	m := metrics.NewMetrics(nil, b.start, b.end)
	countsMap := make(map[string][]float64)
	avgLatsMap := make(map[string][]float64)
	p99LatsMap := make(map[string][]float64)

	for _, r := range records {
		counts, ok := countsMap[r.Type]
		if !ok {
			counts = make([]float64, 0, len(records))
			countsMap[r.Type] = counts
		}
		countsMap[r.Type] = append(counts, r.Count)

		avgLats, ok := avgLatsMap[r.Type]
		if !ok {
			avgLats = make([]float64, 0, len(records))
			avgLatsMap[r.Type] = avgLats
		}
		avgLatsMap[r.Type] = append(avgLats, r.Latency.AvgInMs)

		p99Lats, ok := p99LatsMap[r.Type]
		if !ok {
			p99Lats = make([]float64, 0, len(records))
			p99LatsMap[r.Type] = p99Lats
		}
		p99LatsMap[r.Type] = append(p99Lats, r.Latency.P99InMs)
	}

	results := make([]*Result, 0, 6*len(countsMap))

	for name := range countsMap {
		countJitter := m.CalculateJitter(countsMap[name])
		avgLatJitter := m.CalculateJitter(avgLatsMap[name])
		p99LatJitter := m.CalculateJitter(p99LatsMap[name])
		results = append(results, []*Result{
			{name, "tps-jitter-sd", strconv.FormatFloat(countJitter.Sd, 'f', 2, 64)},
			{name, "tps-jitter-max", strconv.FormatFloat(countJitter.Max, 'f', 2, 64)},

			{name, "avg-lat-jitter-sd", strconv.FormatFloat(avgLatJitter.Sd, 'f', 2, 64)},
			{name, "avg-lat-jitter-max", strconv.FormatFloat(avgLatJitter.Max, 'f', 2, 64)},

			{name, "p99-lat-jitter-sd", strconv.FormatFloat(p99LatJitter.Sd, 'f', 2, 64)},
			{name, "p99-lat-jitter-max", strconv.FormatFloat(p99LatJitter.Max, 'f', 2, 64)},
		}...)
	}

	return results, nil
}
