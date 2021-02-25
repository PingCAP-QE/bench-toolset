import argparse
from analyze.metrics import Connection
from datetime import datetime
from analyze.timeseries import TimeSeries


def parse_args():
    p = argparse.ArgumentParser(description="Analyze jitter from TiDB metrics")

    p.add_argument("--url", type=str, default="http://127.0.0.1:9090", help="Url for connecting Prometheus")
    p.add_argument("--start", type=int, default=0, help="Timestamp in milliseconds")
    p.add_argument("--end", type=int, default=0, help="Timestamp in milliseconds")

    return p.parse_args()


if __name__ == '__main__':
    args = parse_args()
    connect = Connection(args.url)
    start_time = datetime.fromtimestamp(int(args.start / 1000))
    end_time = datetime.fromtimestamp(int(args.end / 1000))
    results = connect.tidb_qps(start_time, end_time)
    time, values = zip(*map(lambda x: (datetime.fromtimestamp(x[0]).strftime("%Y-%m-%d %H:%M:%S"), float(x[1])),
                            results[0]["values"]))
    eval_result = TimeSeries(list(values)).eval(time)
    print('\n'.join(["{:20s}\t{:20s}".format(x, y) for x, y in eval_result.items()]))
