package cmd

import (
	"fmt"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/spf13/cobra"
)

var (
	address        string
	query          string
	begin          int64
	end            int64
	trimSecs       int64
	maxTrimPercent float64
	maxTrimSecs    int64
)

func init() {
	metricsCmd := NewMetricsCommand()

	metricsCmd.PersistentFlags().StringVarP(&address, "address", "u", "", "The host of Prometheus")
	metricsCmd.PersistentFlags().StringVarP(&query, "query", "q", "", "Query of metrics")
	metricsCmd.PersistentFlags().Int64VarP(&begin, "begin", "b", time.Now().UnixMilli()-60000, "Start timestamp in milliseconds")
	metricsCmd.PersistentFlags().Int64VarP(&end, "end", "e", time.Now().UnixMilli(), "End timestamp of statistics")

	metricsCmd.PersistentFlags().Int64Var(&trimSecs, "trim-secs", 180, "trim time range to skip unstable beginning and ending, secs in total")
	metricsCmd.PersistentFlags().Float64Var(&maxTrimPercent, "max-trim-percent", 0.4, "max percentage of trimming")
	metricsCmd.PersistentFlags().Int64Var(&maxTrimSecs, "max-trim-secs", 60*10, "max secs of trimming")

	rootCmd.AddCommand(metricsCmd)
}

func NewMetricsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "metrics",
		Short: "Query metrics from Prometheus",
	}

	command.AddCommand(newJitterCommand())
	command.AddCommand(newAggCommand())

	return command
}

type TimeRange struct {
	begin time.Time
	end   time.Time
}

func (t *TimeRange) trim(trimSecs int64, maxTrimPercent float64, maxTrimSecs int64) {
	totalSecs := t.end.Sub(t.begin).Seconds()
	trimSecsByMaxPercent := int64(totalSecs * maxTrimPercent)
	if trimSecsByMaxPercent < maxTrimSecs {
		maxTrimSecs = trimSecsByMaxPercent
	}
	if trimSecs > maxTrimSecs {
		trimSecs = maxTrimSecs
	}
	trimDuration := time.Second * (time.Duration)(trimSecs/2)
	t.begin = t.begin.Add(trimDuration)
	t.end = t.end.Add(-trimDuration)
}

func newJitterCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "jitter",
		Short: "Calculate jitter for metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			timeRange := &TimeRange{time.UnixMilli(begin), time.UnixMilli(end)}
			timeRange.trim(trimSecs, maxTrimPercent, maxTrimSecs)
			source, err := metrics.NewPrometheus(address)
			if err != nil {
				return err
			}
			result, err := metrics.NewMetrics(source, timeRange.begin, timeRange.end).Jitter(query)
			if err != nil {
				return err
			}

			fmt.Printf("jitter-sd: %f, positive-jitter-max: %f, negative-jitter-max: %f\n", result.Sd, result.PositiveMax.Value, result.NegativeMax.Value)

			return nil
		},
	}

	return command
}

func newAggCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "aggregate",
		Aliases: []string{"agg"},
		Short:   "Calculate average, maximum and minimum for metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := metrics.NewPrometheus(address)
			if err != nil {
				return err
			}
			result, err := metrics.NewMetrics(source, time.UnixMilli(begin), time.UnixMilli(end)).Aggregate(query)
			if err != nil {
				return err
			}

			fmt.Printf("avg: %v, max: %v, min: %v\n", result.Avg, result.Max, result.Min)

			return nil
		},
	}

	return command
}
