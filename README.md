# Bench Toolset
A toolset for running, analyzing benchmarks.

## Usage

### Run benchmark

```
# Run sysbench for one hour
bench-toolset bench sysbench --tables 16 --table-size 10000 --time 1h
# Run tpcc for two hours
bench-toolset bench tpcc --warehouses 1000 --time 2h
```

By default, it uses root@tcp(127.0.0.1:4000)/test as the default dsn address, you can override it by setting below flags:
```
--db string           Database name (default "test")
--host string         Database host (default "127.0.0.1")
--password string     Database password
--port int            Database port (default 4000)
--user string         Database user (default "root")
```

After the benchmark finished, it analyzes the benchmark logs and metrics, outputs the results to terminal and an optional record database, you can use following flags to control behaviours of analyzing:
```
--warmup int          Time for warming up in seconds
--cut-tail int        Time for cutting tail in seconds
--interval int        Sampling interval of logs
--json                Output results to terminal in json
--record-dsn          DSN of database for storing test record
```

See more details with `bench-toolset bench --help`.

### Analyze result
```
# Analyze results of sysbench
bench-toolset analyze log --benchmark sysbench --log sysbench.log --interval 1,2 --warmup 10 --cut-tail 10 -k 2 --percent 10
# Analyze results of tpcc
bench-toolset analyze log --benchmark tpcc --log tpcc.log --interval 1,2 -k 2 --percent 10
```

See more details with `bench-toolset analyze --help`
