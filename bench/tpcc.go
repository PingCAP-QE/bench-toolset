package bench

import (
	"strconv"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

type TpccBench struct {
	load         workload.Workload
	intervalSecs int
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
	m := metrics.NewMetrics(nil, b.start, b.end)
	recordsMap := splitRecordChunks(groupRecords(records), b.intervalSecs)
	results := make([]*Result, 0, 6*len(recordsMap))
	for t, rs := range recordsMap {
		counts := make([]float64, len(rs))
		avgLats := make([]float64, len(rs))
		p99Lats := make([]float64, len(rs))
		countJitter := m.CalculateJitter(counts)
		avgLatJitter := m.CalculateJitter(avgLats)
		p99LatJitter := m.CalculateJitter(p99Lats)
		results = append(results, []*Result{
			{t, "tps-jitter-sd", strconv.FormatFloat(countJitter.Sd, 'f', 2, 64)},
			{t, "tps-jitter-max", strconv.FormatFloat(countJitter.Max, 'f', 2, 64)},

			{t, "avg-lat-jitter-sd", strconv.FormatFloat(avgLatJitter.Sd, 'f', 2, 64)},
			{t, "avg-lat-jitter-max", strconv.FormatFloat(avgLatJitter.Max, 'f', 2, 64)},

			{t, "p99-lat-jitter-sd", strconv.FormatFloat(p99LatJitter.Sd, 'f', 2, 64)},
			{t, "p99-lat-jitter-max", strconv.FormatFloat(p99LatJitter.Max, 'f', 2, 64)},
		}...)
	}

	return results, nil
}

func groupRecords(records []*workload.Record) map[string][]*workload.Record {
	recordsMap := make(map[string][]*workload.Record)
	for _, r := range records {
		_, ok := recordsMap[r.Type]
		if !ok {
			recordsMap[r.Type] = make([]*workload.Record, 0, len(records)/4)
		}
		recordsMap[r.Type] = append(recordsMap[r.Type], r)
	}
	return recordsMap
}

func splitRecordChunks(records map[string][]*workload.Record, chunkSize int) map[string][]*workload.Record {
	res := make(map[string][]*workload.Record)
	for t, rs := range records {
		for i := 0; i < len(rs); i += chunkSize {
			end := i + chunkSize

			if end > len(rs) {
				continue
			}

			sumRecord := &workload.Record{Type: t}
			for _, r := range rs[i:end] {
				sumRecord.Count += r.Count
				sumRecord.AvgLatInMs += r.AvgLatInMs
				if r.P99LatInMs > sumRecord.P99LatInMs {
					sumRecord.P99LatInMs = r.P99LatInMs
				}
			}
			sumRecord.Count /= float64(len(rs))
			sumRecord.AvgLatInMs /= float64(len(rs))
			_, ok := res[t]
			if !ok {
				res[t] = make([]*workload.Record, 0)
				res[t] = append(res[t], sumRecord)
			}
		}
	}
	return res
}
