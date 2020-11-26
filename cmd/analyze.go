package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/5kbpers/bench-toolset/bench"
	"github.com/5kbpers/bench-toolset/workload"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

var (
	benchmark        string
	intervals        string
	promethuesAddr   string
	kNumber          int
	randomPercentage float64
	cutTailSecs      int
)

func init() {
	analyzeCmd := NewAnalyzeCommand()

	analyzeCmd.PersistentFlags().StringVar(&logPath, "log", "", "log path of benchmark")
	analyzeCmd.PersistentFlags().StringVar(&benchmark, "benchmark", "", "benchmark name (tpcc, ycsb, sysbench)")
	analyzeCmd.PersistentFlags().StringVar(&intervals, "interval", "", "interval of analyzing metrics in seconds, separate by ',', eg. 1,5,30")
	analyzeCmd.PersistentFlags().IntVar(&warmupSecs, "warmup", 0, "time of warming up in seconds, will skip the top '${warmup}' records")
	analyzeCmd.PersistentFlags().IntVar(&cutTailSecs, "cut-tail", 0, "time of cutting tail in seconds, will skip the last '${cut-tail}' records")
	analyzeCmd.PersistentFlags().IntVarP(&kNumber, "k-number", "k", 0, "calculate sum of (k-max - k-min)")
	analyzeCmd.PersistentFlags().Float64Var(&randomPercentage, "percent", 0, "percentage for selected number")

	rootCmd.AddCommand(analyzeCmd)
}

func NewAnalyzeCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze logs(support go-tpc, go-ycsb, sysbench) or metrics",
	}
	command.AddCommand(newLogCommand())
	return command
}

func newLogCommand() *cobra.Command {
	command := &cobra.Command{
		Use:  "log",
		Long: "Analyze benchmark(tpcc, ycsb, sysbench) logs, the original report interval must be **1s**",
		RunE: func(cmd *cobra.Command, args []string) error {
			is := make([]int64, 0)
			var records []*workload.Record
			var err error
			for _, s := range strings.Split(intervals, ",") {
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return errors.Trace(err)
				}
				is = append(is, i)
			}
			switch benchmark {
			case "tpcc":
				records, _, err = workload.ParseTpccRecords(logPath)
				if err != nil {
					return err
				}
				if (warmupSecs + cutTailSecs) > len(records) {
					panic("--warmup or --cut-tail maybe too big")
				}
				if len(is) == 0 {
					results := bench.EvalTpccRecords(records, -1, warmupSecs, cutTailSecs, kNumber, randomPercentage)
					for _, r := range results {
						fmt.Printf("%s\t%s\t%s\n", r.Type, r.Name, r.Value)
					}
				} else {
					for _, interval := range is {
						results := bench.EvalTpccRecords(records, int(interval), warmupSecs, cutTailSecs, kNumber, randomPercentage)
						for _, r := range results {
							fmt.Printf("%ds\t%s\t%s\t%s\n", interval, r.Type, r.Name, r.Value)
						}
						fmt.Println("")
					}
				}
			case "sysbench":
				records, _, err = workload.ParseSysbenchRecords(logPath)
				if err != nil {
					return err
				}
				if (warmupSecs + cutTailSecs) > len(records) {
					panic("--warmup or --cut-tail maybe too big")
				}
				fmt.Printf("Found %d records, skip first %d and last %d records.\n\n", len(records), warmupSecs, cutTailSecs)
				if len(is) == 0 {
					results := bench.EvalSysbenchRecords(records, -1, warmupSecs, cutTailSecs, kNumber, randomPercentage)
					for _, r := range results {
						fmt.Printf("%s\t%10s\n", r.Name, r.Value)
					}
				} else {
					for _, interval := range is {
						results := bench.EvalSysbenchRecords(records, int(interval), warmupSecs, cutTailSecs, kNumber, randomPercentage)
						for _, r := range results {
							fmt.Printf("interval=%-4d   %-40s   %15s\n", interval, r.Name, r.Value)
						}
						fmt.Println("")
					}
				}
			default:
				panic("Unsupported benchmark name")
			}
			return nil
		},
	}
	return command
}
