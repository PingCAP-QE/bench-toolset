package bench

import (
	"github.com/5kbpers/bench-toolset/workload"
)

type Benchmark interface {
	Prepare() error
	Run() error
	Results() ([]*Result, error)
}

type Result struct {
	Type  string
	Name  string
	Value string
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
		for _, r := range records[i:end] {
			sumRecord.Count += r.Count
			sumRecord.AvgLatInMs += r.AvgLatInMs
			if r.P99LatInMs > sumRecord.P99LatInMs {
				sumRecord.P99LatInMs = r.P99LatInMs
			}
			if r.P95LatInMs > sumRecord.P95LatInMs {
				sumRecord.P95LatInMs = r.P95LatInMs
			}
		}
		sumRecord.Count /= float64(chunkSize)
		sumRecord.AvgLatInMs /= float64(chunkSize)
		res = append(res, sumRecord)
	}
	return res
}
