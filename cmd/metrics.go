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
	begin   int64
	end     int64
)

func init() {
	metricsCmd := NewMetricsCommand()
	metricsCmd.PersistentFlags().StringVarP(&address, "address", "u", "", "The host of Prometheus")
	metricsCmd.PersistentFlags().StringVarP(&query, "query", "q", "", "Query of metrics")
	metricsCmd.PersistentFlags().Int64VarP(&begin, "begin", "b", time.Now().Unix()-60, "Start of statistics")
	metricsCmd.PersistentFlags().Int64VarP(&end, "end", "e", time.Now().Unix(), "End of statistics")

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
			result, err := metrics.NewMetrics(source, time.Unix(begin, 0), time.Unix(end, 0)).Jitter(query)
			if err != nil {
				return err
			}

			fmt.Printf("jitter-sd: %f, positive-jitter-max: %f, negative-jitter-max: %f\n", result.Sd, result.PositiveMax.Value, result.NegativeMax.Value)

			return nil
		},
	}

	return command
}
