package bench

import (
	"fmt"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

type Benchmark interface {
	Prepare() error
	Run() error
	Results() ([]*Result, []*Result, error)
}

type Result struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
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

func splitRecordChunks(records []*workload.Record, chunkSize int) []*workload.Record {
	res := make([]*workload.Record, 0)
	for i := 0; i < len(records); i += chunkSize {
		end := i + chunkSize

		if end > len(records) {
			continue
		}
		sumRecord := new(workload.Record)
		sumRecord.Payload = make(map[string]interface{})
		for _, r := range records[i:end] {
			sumRecord.Count += r.Count
			sumRecord.AvgLatInMs += r.AvgLatInMs
			for name, value := range r.Payload {
				_, ok := sumRecord.Payload[name]
				if !ok {
					sumRecord.Payload[name] = float64(0)
				}
				if value.(float64) > sumRecord.Payload[name].(float64) {
					sumRecord.Payload[name] = value
				}
			}
		}
		sumRecord.Tag = records[i].Tag
		sumRecord.Count /= float64(chunkSize)
		sumRecord.AvgLatInMs /= float64(chunkSize)
		res = append(res, sumRecord)
	}
	return res
}

func calculateResults(ty string, prefix string, values metrics.TaggedValueSlice, kNumber int, percent float64, unit string) []*Result {
	results := make([]*Result, 0)
	jitter, avg := metrics.CalculateJitter(values, kNumber, percent)
	results = append(results, []*Result{
		{ty, prefix, fmt.Sprintf("%.2f%s", avg, unit)},
		{ty, prefix + "-jitter-sd", fmt.Sprintf("%.2f%%", jitter.Sd*100)},
	}...)
	if len(jitter.PositiveMax.Tag) > 0 {
		results = append(results, &Result{ty, prefix + "-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", jitter.PositiveMax.Value*100, jitter.PositiveMax.Tag)})
	} else {
		results = append(results, &Result{ty, prefix + "-jitter-positive-max", fmt.Sprintf("%.2f%%", jitter.PositiveMax.Value*100)})
	}
	if len(jitter.NegativeMax.Tag) > 0 {
		results = append(results, &Result{ty, prefix + "-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", jitter.NegativeMax.Value*100, jitter.NegativeMax.Tag)})
	} else {
		results = append(results, &Result{ty, prefix + "-jitter-negative-max", fmt.Sprintf("%.2f%%", jitter.NegativeMax.Value*100)})
	}

	if kNumber > 0 {
		results = append(results, &Result{
			"",
			fmt.Sprintf("%s-jitter-kmean(k=%d)", prefix, kNumber),
			fmt.Sprintf("%.2f%%", jitter.KMean*100),
		})
	}
	if percent > 0 {
		results = append(results, &Result{
			"",
			fmt.Sprintf("%s-jitter-rand-percent(p=%.2f%%)", prefix, percent),
			fmt.Sprintf("%.2f%%", jitter.Random*100),
		})
	}
	return results
}
