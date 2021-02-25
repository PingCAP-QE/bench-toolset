# bench-toolset
## Usage
### Analyze
#### Analyze Log
```bash
python3 analyze_log.py sysbench --log sysbench.log
```

#### Analyze TiDB QPS
```bash
python3 analyze_tidb_metrics.py --url="http://127.0.0.1:9090/" --start 1614240842504 --end 1614242513038
```