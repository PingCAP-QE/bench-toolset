package bench

import (
	"fmt"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

type TpccBench struct {
	load         workload.Workload
	intervalSecs int
	warmupSecs   int
	cutTailSecs  int
	start        time.Time
	end          time.Time
}

func NewTpccBench(load workload.Tpcc, intervalSecs int, warmupSecs int, cutTailSecs int) *TpccBench {
	return &TpccBench{
		load:         &load,
		intervalSecs: intervalSecs,
		warmupSecs:   warmupSecs,
		cutTailSecs:  cutTailSecs,
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

func (b *TpccBench) Results() ([]*Result, []*Result, error) {
	records, summaryRecords, err := b.load.Records()
	if err != nil {
		return nil, nil, err
	}
	results := EvalTpccRecords(records, b.intervalSecs, b.warmupSecs, b.cutTailSecs, 0, 0)
	summaryResults := EvalTpccSummaryRecord(summaryRecords)
	return results, summaryResults, nil
}

func EvalTpccRecords(records []*workload.Record, intervalSecs int, warmupSecs int, cutTailSecs int, kNumber int, percent float64) []*Result {
	recordsMap := groupRecords(records)
	if intervalSecs > 0 {
		for t, rs := range recordsMap {
			recordsMap[t] = splitRecordChunks(rs[warmupSecs:len(rs)-cutTailSecs], intervalSecs)
			fmt.Printf("Aggregate %s records with interval %d, got %d records.\n", t, intervalSecs, len(recordsMap[t]))
		}
	}
	results := make([]*Result, 0, 6*len(recordsMap))
	for t, rs := range recordsMap {
		counts := make(metrics.TaggedValueSlice, len(rs))
		avgLats := make(metrics.TaggedValueSlice, len(rs))
		payloads := make(map[string]metrics.TaggedValueSlice)
		for i, r := range rs {
			counts[i] = metrics.WithTag(r.Count, r.Tag)
			avgLats[i] = metrics.WithTag(r.AvgLatInMs, r.Tag)

			for p, v := range r.Payload {
				_, ok := payloads[p]
				if !ok {
					payloads[p] = make(metrics.TaggedValueSlice, 0, len(rs))
				}
				payloads[p] = append(payloads[p], metrics.WithTag(v.(float64), r.Tag))
			}
		}
		results = append(results, calculateResults(t, "tps", counts, kNumber, percent, "")...)
		results = append(results, calculateResults(t, "avg-lat", avgLats, kNumber, percent, "ms")...)
		for p, vs := range payloads {
			results = append(results, calculateResults(t, p, vs, kNumber, percent, "ms")...)
		}
	}
	return results
}

func EvalTpccSummaryRecord(records []*workload.Record) []*Result {
	results := make([]*Result, 0)
	for _, item := range records {
		tpm := item.Payload["tpm"]
		results = append(results, &Result{
			Type:  item.Type,
			Name:  "tpm",
			Value: fmt.Sprintf("%.2f", tpm),
		})
	}
	return results
}
