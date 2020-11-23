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
	results := EvalTpccRecords(records, b.intervalSecs, b.warmupSecs, 0, 0)
	return results, nil
}

func EvalTpccRecords(records []*workload.Record, intervalSecs int, warmupSecs int, kNumber int, percent float64) []*Result {
	recordsMap := groupRecords(records)
	if intervalSecs > 0 {
		for t, rs := range recordsMap {
			recordsMap[t] = splitRecordChunks(rs[warmupSecs:], intervalSecs)
		}
	}
	results := make([]*Result, 0, 6*len(recordsMap))
	for t, rs := range recordsMap {
		counts := make([]float64, len(rs))
		avgLats := make([]float64, len(rs))
		p99Lats := make([]float64, len(rs))
		for i, r := range rs {
			counts[i] = r.Count
			avgLats[i] = r.AvgLatInMs
			p99Lats[i] = r.P99LatInMs
		}
		countJitter, countSum := metrics.CalculateJitter(counts, kNumber, percent)
		avgLatJitter, avgLatSum := metrics.CalculateJitter(avgLats, kNumber, percent)
		p99LatJitter, _ := metrics.CalculateJitter(p99Lats, kNumber, percent)
		results = append(results, []*Result{
			{t, "tps-jitter-sd", fmt.Sprintf("%.2f%%", countJitter.Sd*100)},
			{t, "tps-jitter-max", fmt.Sprintf("%.2f%%", countJitter.Max*100)},

			{t, "avg-lat-jitter-sd", fmt.Sprintf("%.2f%%", avgLatJitter.Sd*100)},
			{t, "avg-lat-jitter-max", fmt.Sprintf("%.2f%%", avgLatJitter.Max*100)},

			{t, "p99-lat-jitter-sd", fmt.Sprintf("%.2f%%", p99LatJitter.Sd*100)},
			{t, "p99-lat-jitter-max", fmt.Sprintf("%.2f%%", p99LatJitter.Max*100)},

			{t, "avg-tps", strconv.FormatFloat(countSum/float64(len(rs)), 'f', 2, 64)},
			{t, "avg-lat-in-ms", strconv.FormatFloat(avgLatSum/float64(len(rs)), 'f', 2, 64)},
		}...)
	}
	return results
}
