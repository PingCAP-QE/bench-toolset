package cmd

import (
	"database/sql"
	"time"

	"github.com/5kbpers/stability_bench/bench"
	"github.com/5kbpers/stability_bench/workload"
	"github.com/go-sql-driver/mysql"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	host        string
	user        string
	port        uint64
	db          string
	threads     uint64
	logPath     string
	recordDbDsn string
	runTime     time.Duration

	recordDb *sql.DB

	sysbenchTables    uint64
	sysbenchTableSize uint64
	sysbenchName      string

	tpccWareHouses uint64
)

func init() {
	benchCmd := NewBenchCommand()

	benchCmd.PersistentFlags().StringVar(&host, "host", "127.0.0.1", "host of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&user, "user", "root", "username of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&db, "db", "test", "database of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&logPath, "log", "", "log path of workload")
	benchCmd.PersistentFlags().Uint64Var(&port, "port", 4000, "port of tidb cluster")
	benchCmd.PersistentFlags().Uint64Var(&threads, "threads", 16, "port of tidb cluster")
	benchCmd.PersistentFlags().StringVar(&recordDbDsn, "record-dsn", "", "Dsn of database for storing test record")

	rootCmd.AddCommand(benchCmd)
}

func NewBenchCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "bench",
		Short: "Run benchmarks for stability test",
	}

	command.AddCommand(newGcCommand())

	return command
}

func newGcCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "tpcc",
		Short: "Run Tpc-C workload",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var benchId int64
			if len(recordDbDsn) != 0 {
				config, err := mysql.ParseDSN(recordDbDsn)
				if err != nil {
					return errors.Trace(err)
				}
				log.Info("Parse record database DSN", zap.Reflect("config", config))
				log.Info("Connect to record database...")
				recordDb, err = sql.Open("mysql", recordDbDsn)
				if err != nil {
					return errors.Trace(err)
				}
				err = recordDb.Ping()
				if err != nil {
					return errors.Trace(err)
				}
				var rs sql.Result
				log.Info("Create a record for this benchmark...")
				rs, err = recordDb.Exec(`INSERT INTO bench_info (name) VALUES ("gc")`)
				if err != nil {
					return errors.Trace(err)
				}
				log.Info("Get benchmark id...", zap.Int64("id", benchId))
				benchId, err = rs.LastInsertId()
				if err != nil {
					return errors.Trace(err)
				}
			}

			load := &workload.Tpcc{
				WareHouses: tpccWareHouses,
				Db:         db,
				Host:       host,
				Port:       port,
				User:       user,
				Threads:    threads,
				Time:       runTime,
				LogPath:    logPath,
			}
			b := bench.NewGcBench(load)
			log.Info("Prepare benchmark...")
			err = b.Prepare()
			if err != nil {
				return errors.Trace(err)
			}
			log.Info("Start to run benchmark...")
			err = b.Run()
			if err != nil {
				return errors.Trace(err)
			}
			results, err := b.Results()
			if err != nil {
				return errors.Trace(err)
			}
			log.Info("Benchmark done, save results to record database...", zap.Reflect("results", results))
			if recordDb != nil {
				for _, rs := range results {
					_, err = recordDb.Exec("INSERT INTO bench_result values (?, ?, ?)", benchId, rs.Name, rs.Value)
					if err != nil {
						return errors.Trace(err)
					}
				}
			}
			return nil
		},
	}

	command.PersistentFlags().Uint64Var(&tpccWareHouses, "warehouse", 16, "table count of sysbench workload")
	command.PersistentFlags().DurationVar(&runTime, "time", time.Hour, "running time of workload")

	return command
}
