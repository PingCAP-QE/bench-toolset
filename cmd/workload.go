package cmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/5kbpers/stability_bench/bench"
	"github.com/5kbpers/stability_bench/workload"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var (
	host        string
	user        string
	port        uint64
	db          string
	threads     uint64
	logPath     string
	recordDbDsn string

	recordDb *sql.DB

	sysbenchTables    uint64
	sysbenchTableSize uint64
	sysbenchTime      time.Duration
)

func init() {
	workloadCmd := NewWorkloadCommand()

	workloadCmd.PersistentFlags().StringVar(&host, "host", "localhost", "host of tidb cluster")
	workloadCmd.PersistentFlags().StringVar(&user, "user", "root", "username of tidb cluster")
	workloadCmd.PersistentFlags().StringVar(&db, "db", "test", "database of tidb cluster")
	workloadCmd.PersistentFlags().StringVar(&logPath, "log", "", "log path of workload")
	workloadCmd.PersistentFlags().Uint64Var(&port, "port", 4000, "port of tidb cluster")
	workloadCmd.PersistentFlags().Uint64Var(&threads, "threads", 16, "port of tidb cluster")
	workloadCmd.PersistentFlags().StringVar(&recordDbDsn, "record-dsn", "", "Dsn of database for storing test record")

	rootCmd.AddCommand(workloadCmd)
}

func NewWorkloadCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "workload",
		Short: "Run workloads for stability test",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			recordDb, err = sql.Open("mysql", recordDbDsn)
			return err
		},
	}

	command.AddCommand(newGcWorkloadCommand())

	return command
}

func newGcWorkloadCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "gc",
		Short: "Run GC workload",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var benchId int64
			if recordDb != nil {
				var rs sql.Result
				rs, err = recordDb.Exec(`INSERT INTO bench_info ("name") VALUES ("gc")`)
				if err != nil {
					return err
				}
				benchId, err = rs.LastInsertId()
				if err != nil {
					return err
				}
			}

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
			err = b.Run()
			if err != nil {
				return err
			}
			results, err := b.Results()
			if err != nil {
				return err
			}
			fmt.Println(results)

			if recordDb != nil {
				for _, rs := range results {
					_, err = recordDb.Exec("INSERT INTO bench_record values (?, ?, ?)", benchId, rs.Name, rs.Value)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	command.PersistentFlags().Uint64Var(&sysbenchTables, "tables", 16, "table count of sysbench workload")
	command.PersistentFlags().Uint64Var(&sysbenchTableSize, "size", 100000, "table size of sysbench workload")
	command.PersistentFlags().DurationVar(&sysbenchTime, "time", time.Hour, "running time of sysbench workload")

	return command
}
