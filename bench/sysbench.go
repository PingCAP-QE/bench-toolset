package bench

import (
	"fmt"
	"strconv"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/5kbpers/bench-toolset/workload"
)

func EvalSysbenchRecords(records []*workload.Record, intervalSecs int, warmupSecs int, cutTailSecs int, kNumber int, percent float64) []*Result {
	recordsMap := groupRecords(records)
	if intervalSecs > 0 {
		for t, rs := range recordsMap {
			recordsMap[t] = splitRecordChunks(rs[warmupSecs:len(rs)-cutTailSecs], intervalSecs)
			fmt.Printf("Aggregate records with interval %d, got %d records.\n", intervalSecs, len(recordsMap[t]))
		}
	}
	results := make([]*Result, 0, 6*len(recordsMap))
	for _, rs := range recordsMap {
		counts := make([]float64, len(rs))
		avgLats := make([]float64, len(rs))
		p95Lats := make([]float64, 0, len(rs))
		p99Lats := make([]float64, 0, len(rs))
		for i, r := range rs {
			counts[i] = r.Count
			avgLats[i] = r.AvgLatInMs
			if r.P95LatInMs > 0 {
				p95Lats = append(p95Lats, r.P95LatInMs)
			}
			if r.P99LatInMs > 0 {
				p99Lats = append(p99Lats, r.P99LatInMs)
			}
		}
		countJitter, countAvg := metrics.CalculateJitter(counts, kNumber, percent)
		results = append(results, []*Result{
			{"", "tps-jitter-sd", fmt.Sprintf("%.2f%%", countJitter.Sd*100)},
			{"", "tps-jitter-max", fmt.Sprintf("%.2f%%", countJitter.Max*100)},
			{"", "avg-tps", strconv.FormatFloat(countAvg, 'f', 2, 64)},
		}...)
		if kNumber > 0 {
			results = append(results, &Result{
				"",
				fmt.Sprintf("tps-jitter-kmean(k=%d)", kNumber),
				fmt.Sprintf("%.2f%%", countJitter.KMean*100),
			})
		}
		if percent > 0 {
			results = append(results, &Result{
				"",
				fmt.Sprintf("count-jitter-rand-percent(p=%.2f%%)", percent),
				fmt.Sprintf("%.2f%%", countJitter.Random*100),
			})
		}

		avgLatJitter, avgLatAvg := metrics.CalculateJitter(avgLats, kNumber, percent)
		results = append(results, []*Result{
			{"", "avg-lat-jitter-sd", fmt.Sprintf("%.2f%%", avgLatJitter.Sd*100)},
			{"", "avg-lat-jitter-max", fmt.Sprintf("%.2f%%", avgLatJitter.Max*100)},
			{"", "avg-lat-in-ms", strconv.FormatFloat(avgLatAvg, 'f', 2, 64)},
		}...)
		if kNumber > 0 {
			results = append(results, &Result{
				"",
				fmt.Sprintf("avg-lat-jitter-kmean(k=%d)", kNumber),
				fmt.Sprintf("%.2f%%", avgLatJitter.KMean*100),
			})
		}
		if percent > 0 {
			results = append(results, &Result{
				"",
				fmt.Sprintf("avg-lat-jitter-rand-percent(p=%.2f%%)", percent),
				fmt.Sprintf("%.2f%%", avgLatJitter.Random*100),
			})
		}

		if len(p95Lats) > 0 {
			p95LatJitter, p95LatAvg := metrics.CalculateJitter(p95Lats, kNumber, percent)
			results = append(results, []*Result{
				{"", "p95-lat-jitter-sd", fmt.Sprintf("%.2f%%", p95LatJitter.Sd*100)},
				{"", "p95-lat-jitter-max", fmt.Sprintf("%.2f%%", p95LatJitter.Max*100)},
				{"", "p95-lat-in-ms", strconv.FormatFloat(p95LatAvg, 'f', 2, 64)},
			}...)
			if kNumber > 0 {
				results = append(results, &Result{
					"",
					fmt.Sprintf("p95-lat-jitter-kmean(k=%d)", kNumber),
					fmt.Sprintf("%.2f%%", p95LatJitter.KMean*100),
				})
			}
			if percent > 0 {
				results = append(results, &Result{
					"",
					fmt.Sprintf("p95-lat-jitter-rand-percent(p=%.2f%%)", percent),
					fmt.Sprintf("%.2f%%", p95LatJitter.Random*100),
				})
			}
		}

		if len(p99Lats) > 0 {
			p99LatJitter, p99LatAvg := metrics.CalculateJitter(p99Lats, kNumber, percent)
			results = append(results, []*Result{
				{"", "p99-lat-jitter-sd", fmt.Sprintf("%.2f%%", p99LatJitter.Sd*100)},
				{"", "p99-lat-jitter-max", fmt.Sprintf("%.2f%%", p99LatJitter.Max*100)},
				{"", "p99-lat-in-ms", strconv.FormatFloat(p99LatAvg, 'f', 2, 64)},
			}...)
			if kNumber > 0 {
				results = append(results, &Result{
					"",
					fmt.Sprintf("p99-lat-jitter-kmean(k=%d)", kNumber),
					fmt.Sprintf("%.2f%%", p99LatJitter.KMean*100),
				})
			}
			if percent > 0 {
				results = append(results, &Result{
					"",
					fmt.Sprintf("p99-lat-jitter-rand-percent(p=%.2f%%)", percent),
					fmt.Sprintf("%.2f%%", p99LatJitter.Random*100),
				})
			}
		}
	}
	return results
}
