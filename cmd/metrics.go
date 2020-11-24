package cmd

import (
	"fmt"
	"time"

	"github.com/5kbpers/bench-toolset/metrics"
	"github.com/spf13/cobra"
)

var (
	address string
	query   string

	interval time.Duration
)

func init() {
	metricsCmd := NewMetricsCommand()
	metricsCmd.PersistentFlags().StringVarP(&address, "address", "u", "", "The host of Prometheus")
	metricsCmd.PersistentFlags().StringVarP(&query, "query", "q", "", "Query of metrics")

	rootCmd.AddCommand(metricsCmd)
}

func NewMetricsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "metrics",
		Short: "Query metrics from Prometheus",
	}

	command.AddCommand(newJitterCommand())

	return command
}

func newJitterCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "jitter",
		Short: "Calculate jitter for metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := metrics.NewPrometheus(address)
			if err != nil {
				return err
			}
			now := time.Now()
			result, err := metrics.NewMetrics(source, now, now.Add(interval)).Jitter(query)
			if err != nil {
				return err
			}

			fmt.Printf("jitter-sd: %f, positive-jitter-max: %f, negative-jitter-max: %f\n", result.Sd, result.PositiveMax.Value, result.NegativeMax.Value)

			return nil
		},
	}

	command.PersistentFlags().DurationVar(&interval, "time", time.Minute*10, "Time of fetching metrics")
	return command
}
