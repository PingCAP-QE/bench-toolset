package bench

import (
	"fmt"
	"strconv"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

type TpccBench struct {
	load         workload.Workload
	intervalSecs int
	warmupSecs   int
	start        time.Time
	end          time.Time
}

func NewTpccBench(load workload.Workload, intervalSecs int) *TpccBench {
	return &TpccBench{
		load:         load,
		intervalSecs: intervalSecs,
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
	results := EvalTpccRecords(records, b.intervalSecs, b.warmupSecs, 0, 0, 0)
	return results, nil
}

func EvalTpccRecords(records []*workload.Record, intervalSecs int, warmupSecs int, cutTailSecs int, kNumber int, percent float64) []*Result {
	recordsMap := groupRecords(records)
	if intervalSecs > 0 {
		for t, rs := range recordsMap {
			recordsMap[t] = splitRecordChunks(rs[warmupSecs:], intervalSecs)
		}
	}
	results := make([]*Result, 0, 6*len(recordsMap))
	for t, rs := range recordsMap {
		counts := make(metrics.TaggedValueSlice, len(rs))
		avgLats := make(metrics.TaggedValueSlice, len(rs))
		p95Lats := make(metrics.TaggedValueSlice, 0, len(rs))
		p99Lats := make(metrics.TaggedValueSlice, 0, len(rs))
		for i, r := range rs {
			counts[i] = metrics.WithTag(r.Count, r.Tag)
			avgLats[i] = metrics.WithTag(r.AvgLatInMs, r.Tag)
			if r.P95LatInMs > 0 {
				p95Lats = append(p95Lats, metrics.WithTag(r.P95LatInMs, r.Tag))
			}
			if r.P99LatInMs > 0 {
				p99Lats = append(p99Lats, metrics.WithTag(r.P99LatInMs, r.Tag))
			}
		}
		countJitter, countSum := metrics.CalculateJitter(counts, kNumber, percent)
		avgLatJitter, avgLatSum := metrics.CalculateJitter(avgLats, kNumber, percent)
		p99LatJitter, _ := metrics.CalculateJitter(p99Lats, kNumber, percent)
		results = append(results, []*Result{
			{t, "tps-jitter-sd", fmt.Sprintf("%.2f%%", countJitter.Sd*100)},
			{t, "tps-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", countJitter.PositiveMax.Value*100, countJitter.PositiveMax.Tag)},
			{t, "tps-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", countJitter.NegativeMax.Value*100, countJitter.NegativeMax.Tag)},

			{t, "avg-lat-jitter-sd", fmt.Sprintf("%.2f%%", avgLatJitter.Sd*100)},
			{t, "avg-lat-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", avgLatJitter.PositiveMax.Value*100, avgLatJitter.PositiveMax.Tag)},
			{t, "avg-lat-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", avgLatJitter.NegativeMax.Value*100, avgLatJitter.NegativeMax.Tag)},

			{t, "p99-lat-jitter-sd", fmt.Sprintf("%.2f%%", p99LatJitter.Sd*100)},
			{t, "p99-lat-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", p99LatJitter.PositiveMax.Value*100, p99LatJitter.PositiveMax.Tag)},
			{t, "p99-lat-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", p99LatJitter.NegativeMax.Value*100, p99LatJitter.NegativeMax.Tag)},

			{t, "avg-tps", strconv.FormatFloat(countSum/float64(len(rs)), 'f', 2, 64)},
			{t, "avg-lat-in-ms", strconv.FormatFloat(avgLatSum/float64(len(rs)), 'f', 2, 64)},
		}...)
	}
	return results
}
