package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/5kbpers/bench-toolset/bench"
	"github.com/5kbpers/bench-toolset/workload"
	"github.com/go-sql-driver/mysql"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	host         string
	user         string
	port         uint64
	db           string
	threads      uint64
	logPath      string
	recordDbDsn  string
	runTime      time.Duration
	intervalSecs int
	warmupSecs   int

	brArgs         []string
	prometheusAddr string

	sysbenchTables    uint64
	sysbenchTableSize uint64
	sysbenchName      string

	tpccWareHouses uint64

	outputJson bool
)

var (
	benchId        int64
	recordDbConn   *sql.Conn
	results        []*bench.Result
	summaryResults []*bench.Result
)

func init() {
	benchCmd := NewBenchCommand()

	benchCmd.PersistentFlags().StringVar(&host, "host", "127.0.0.1", "host of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&user, "user", "root", "username of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&db, "db", "test", "database of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&logPath, "log", "", "log path of workload")
	benchCmd.PersistentFlags().Uint64Var(&port, "port", 4000, "port of tidb cluster")
	benchCmd.PersistentFlags().Uint64Var(&threads, "threads", 16, "port of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&recordDbDsn, "record-dsn", "", "dsn of database for storing test record")
	benchCmd.PersistentFlags().IntVar(&intervalSecs, "interval", -1, "interval of metrics in seconds")
	benchCmd.PersistentFlags().StringArrayVar(&brArgs, "br-args", []string{}, "args of br restore")
	benchCmd.PersistentFlags().StringVar(&prometheusAddr, "prometheus", "", "addr of prometheus")
	benchCmd.PersistentFlags().IntVar(&warmupSecs, "warmup", 0, "time of warming up in seconds, will skip the top '${warmup}' records ")
	benchCmd.PersistentFlags().IntVar(&cutTailSecs, "cut-tail", 0, "time of cutting tail in seconds, will skip the last '${cut-tail}' records")
	benchCmd.PersistentFlags().BoolVar(&outputJson, "json", false, "output results in json")

	rootCmd.AddCommand(benchCmd)
}

func NewBenchCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "bench",
		Short: "Run benchmarks",
	}

	command.AddCommand(newTpccCommand())
	command.AddCommand(newSysbenchCommand())

	return command
}

func maybeCreateTestRecord(name string) error {
	if len(recordDbDsn) != 0 {
		config, err := mysql.ParseDSN(recordDbDsn)
		if err != nil {
			return errors.Trace(err)
		}
		log.Info("Parse record database DSN", zap.Reflect("config", config))
		recordDb, err := sql.Open("mysql", recordDbDsn)
		if err != nil {
			return errors.Trace(err)
		}
		log.Info("Connect to record database...")
		recordDbConn, err = recordDb.Conn(context.Background())
		if err != nil {
			return errors.Trace(err)
		}
		var rs sql.Result
		log.Info("Create a record for this benchmark...")
		rs, err = recordDb.Exec(`INSERT INTO bench_info (name) VALUES (?)`, name)
		if err != nil {
			return errors.Trace(err)
		}
		benchId, err = rs.LastInsertId()
		if err != nil {
			return errors.Trace(err)
		}
		log.Info("Get benchmark id", zap.Int64("id", benchId))
	}
	return nil
}

func maybeSaveTestResults() error {
	if recordDbConn != nil {
		log.Info("Save results to record database...")
		_, err := recordDbConn.ExecContext(context.Background(), "UPDATE bench_info SET end_time=NOW() where id=?", benchId)
		if err != nil {
			return errors.Trace(err)
		}
		for _, rs := range results {
			_, err = recordDbConn.ExecContext(context.Background(), "INSERT INTO bench_result values (?, ?, ?, ?)", benchId, rs.Name, rs.Value, rs.Type)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	if outputJson {
		var totalResults = append(results, summaryResults...)
		j, err := json.Marshal(totalResults)
		if err != nil {
			return errors.Trace(err)
		}
		fmt.Println(string(j))
	}
	return nil
}

func newTpccCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "tpcc",
		Short: "Run Tpc-C workload",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return maybeCreateTestRecord("tpcc")
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return maybeSaveTestResults()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			load := workload.Tpcc{
				WareHouses: tpccWareHouses,
				Db:         db,
				Host:       host,
				Port:       port,
				User:       user,
				Threads:    threads,
				Time:       runTime,
				LogPath:    logPath,
			}
			b := bench.NewTpccBench(load, intervalSecs, warmupSecs, cutTailSecs)
			log.Info("Prepare benchmark...")
			var err error
			if len(brArgs) > 0 {
				log.Info("Run BR restore...")
				err = runBrRestore(brArgs)
				if err != nil {
					return errors.Trace(err)
				}
			} else {
				err = b.Prepare()
				if err != nil {
					return errors.Trace(err)
				}
			}
			log.Info("Start to run benchmark...")
			err = b.Run()
			if err != nil {
				return errors.Trace(err)
			}
			results, summaryResults, err = b.Results()
			if err != nil {
				return errors.Trace(err)
			}
			log.Info("Benchmark done", zap.Reflect("results", results))
			return nil
		},
	}

	command.PersistentFlags().Uint64Var(&tpccWareHouses, "warehouses", 16, "warehouses count of tpcc workload")
	command.PersistentFlags().DurationVar(&runTime, "time", time.Hour, "running time of workload")

	return command
}

func newSysbenchCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "sysbench",
		Short: "Run Sysbench workload",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return maybeCreateTestRecord("sysbench")
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return maybeSaveTestResults()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			load := workload.Sysbench{
				Name:      sysbenchName,
				Tables:    sysbenchTables,
				TableSize: sysbenchTableSize,
				Db:        db,
				Host:      host,
				Port:      port,
				User:      user,
				Threads:   threads,
				Time:      runTime,
				LogPath:   logPath,
			}
			b := bench.NewSysbenchBench(load, intervalSecs, warmupSecs, cutTailSecs)
			log.Info("Prepare benchmark...")
			var err error
			if len(brArgs) > 0 {
				log.Info("Run BR restore...")
				err = runBrRestore(brArgs)
				if err != nil {
					return errors.Trace(err)
				}
			} else {
				err = b.Prepare()
				if err != nil {
					return errors.Trace(err)
				}
			}
			log.Info("Start to run benchmark...")
			err = b.Run()
			if err != nil {
				return errors.Trace(err)
			}
			results, _, err = b.Results()
			if err != nil {
				return errors.Trace(err)
			}
			log.Info("Benchmark done", zap.Reflect("results", results))
			return nil
		},
	}

	command.PersistentFlags().Uint64Var(&sysbenchTables, "tables", 16, "table count of sysbench workload")
	command.PersistentFlags().Uint64Var(&sysbenchTableSize, "table-size", 10000, "table size of sysbench workload")
	command.PersistentFlags().StringVar(&sysbenchName, "name", "oltp_update_index", "test name of sysbench workload")
	command.PersistentFlags().DurationVar(&runTime, "time", time.Hour, "running time of workload")

	return command
}
