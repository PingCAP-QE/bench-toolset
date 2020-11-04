package cmd

import (
	"fmt"
	"time"

	"github.com/5kbpers/stability_bench/bench"
	"github.com/5kbpers/stability_bench/workload"
	"github.com/spf13/cobra"
)

var (
	host    string
	user    string
	port    uint64
	db      string
	threads uint64
	logPath string

	sysbenchTables    uint64
	sysbenchTableSize uint64
	sysbenchTime      time.Duration
)

func init() {
	workloadCmd := NewWorkloadCommand()

	workloadCmd.PersistentFlags().StringVarP(&host, "host", "o", "localhost", "host of tidb cluster")
	workloadCmd.PersistentFlags().StringVarP(&user, "user", "u", "root", "username of tidb cluster")
	workloadCmd.PersistentFlags().StringVarP(&db, "db", "d", "test", "database of tidb cluster")
	workloadCmd.PersistentFlags().StringVarP(&logPath, "log", "l", "", "log path of workload")
	workloadCmd.PersistentFlags().Uint64VarP(&port, "port", "p", 4000, "port of tidb cluster")
	workloadCmd.PersistentFlags().Uint64VarP(&threads, "threads", "t", 16, "port of tidb cluster")

	rootCmd.AddCommand(workloadCmd)
}

func NewWorkloadCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "workload",
		Short: "Run workloads for stability test",
	}

	command.AddCommand(newGcWorkloadCommand())

	return command
}

func newGcWorkloadCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "gc",
		Short: "Run GC workload",
		RunE: func(cmd *cobra.Command, args []string) error {
			load := &workload.Sysbench{
				Name:           "oltp_update_index",
				Host:           host,
				User:           user,
				Port:           port,
				Db:             db,
				Threads:        threads,
				Tables:         sysbenchTables,
				TableSize:      sysbenchTableSize,
				Time:           sysbenchTime,
				ReportInterval: time.Second * 10,
				LogPath:        logPath,
			}
			b := bench.NewGcBench(load)
			err := b.Run()
			if err != nil {
				return err
			}
			report, err := b.Report()
			if err != nil {
				return err
			}
			fmt.Println(report)
			return nil
		},
	}

	command.PersistentFlags().Uint64Var(&sysbenchTables, "tables", 16, "table count of sysbench workload")
	command.PersistentFlags().Uint64Var(&sysbenchTableSize, "size", 100000, "table size of sysbench workload")
	command.PersistentFlags().DurationVar(&sysbenchTime, "time", time.Hour, "running time of sysbench workload")

	return command
}
