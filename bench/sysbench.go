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
		countJitter, countAvg := metrics.CalculateJitter(counts, kNumber, percent)
		results = append(results, []*Result{
			{"", "tps-jitter-sd", fmt.Sprintf("%.2f%%", countJitter.Sd*100)},
			{"", "tps-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", countJitter.PositiveMax.Value*100, countJitter.PositiveMax.Tag)},
			{"", "tps-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", countJitter.NegativeMax.Value*100, countJitter.NegativeMax.Tag)},
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
				fmt.Sprintf("tps-jitter-rand-percent(p=%.2f%%)", percent),
				fmt.Sprintf("%.2f%%", countJitter.Random*100),
			})
		}

		avgLatJitter, avgLatAvg := metrics.CalculateJitter(avgLats, kNumber, percent)
		results = append(results, []*Result{
			{"", "avg-lat-jitter-sd", fmt.Sprintf("%.2f%%", avgLatJitter.Sd*100)},
			{"", "avg-lat-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", avgLatJitter.PositiveMax.Value*100, avgLatJitter.PositiveMax.Tag)},
			{"", "avg-lat-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", avgLatJitter.NegativeMax.Value*100, avgLatJitter.NegativeMax.Tag)},
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
				{"", "p95-lat-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", p95LatJitter.PositiveMax.Value*100, p95LatJitter.PositiveMax.Tag)},
				{"", "p95-lat-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", p95LatJitter.NegativeMax.Value*100, p95LatJitter.NegativeMax.Tag)},
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
				{"", "p99-lat-jitter-positive-max", fmt.Sprintf("%.2f%% in %s", p99LatJitter.PositiveMax.Value*100, p99LatJitter.PositiveMax.Tag)},
				{"", "p99-lat-jitter-negative-max", fmt.Sprintf("%.2f%% in %s", p99LatJitter.NegativeMax.Value*100, p99LatJitter.NegativeMax.Tag)},
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
