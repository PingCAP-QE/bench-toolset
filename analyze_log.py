import argparse
import pathlib
from bench.sysbench import Sysbench


def parse_args():
    p = argparse.ArgumentParser(description="Analyze log of benchmark result, support sysbench, go-tpc")

    p.add_argument("name", help="Benchmark name", choices=["sysbench"])
    p.add_argument("--log", type=pathlib.Path, help="Log path")
    p.add_argument("--interval", type=int, default=10, help="Output interval of log file in seconds")
    p.add_argument("--warmup", type=int, default=0, help="Warmup time in seconds")
    p.add_argument("--tail", type=int, default=0, help="Tail time in seconds")
    p.add_argument("--group", type=int, default=10, help="Group records into chunks in the size of 'group'/'interval'")

    return p.parse_args()


if __name__ == '__main__':
    args = parse_args()
    Sysbench.set_conf("log-path", args.log)
    Sysbench.parse()
    warmup = int(args.warmup / args.interval)
    tail = int(args.tail / args.interval)
    records = Sysbench.grouped_records(int(args.group / args.interval), warmup, Sysbench.record_count() - tail)
    print(Sysbench.eval_records_string(records), "\n")
