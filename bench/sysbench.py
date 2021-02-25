import re
import os
from analyze.timeseries import TimeSeries

recordRegex = re.compile(
    r'\[\s(\d+s)\s]\sthds:\s\d+\stps:\s([\d.]+)\sqps:\s[\d.]+\s[()\w/:\s\d.]+\slat\s\(ms,\d+%\):\s([\d.]+)')
tpsRegex = re.compile(r'\n\s+transactions:\s+([\d.]+)\s*\(([\d.]+)\s+per\s+sec\.\)')
qpsRegex = re.compile(r'\n\s+queries:\s+([\d.]+)\s*\(([\d.]+)\s+per\s+sec\.\)')
minLatRegex = re.compile(r'\n\s+min:\s+([\d.]+)')
avgLatRegex = re.compile(r'\n\s+avg:\s+([\d.]+)')
maxLatRegex = re.compile(r'\n\s+max:\s+([\d.]+)')
p99LatRegex = re.compile(r'\n\s+99th percentile:\s+([\d.]+)')


class Sysbench:
    __conf = {
        "mysql-host": "",
        "mysql-user": "",
        "mysql-port": 0,
        "mysql-db": "",
        "tables": 0,
        "table-size": 0,
        "name": "",
        "threads": 0,
        "time": 0,
        "log-path": "",
    }
    __records = {
        "time": TimeSeries([]),
        "tps": TimeSeries([]),
        "p95_lat_ms": TimeSeries([]),
        "avg_lat_ms": TimeSeries([]),
    }
    __result = []

    @staticmethod
    def set_conf(name, value):
        if name in Sysbench.__conf.keys():
            Sysbench.__conf[name] = value
        else:
            raise NameError("config item not accepted in sysbench")

    @staticmethod
    def records():
        return Sysbench.__records

    @staticmethod
    def record_count():
        return len(Sysbench.__records["time"])

    @staticmethod
    def prepare():
        args = Sysbench.__build_args__()
        os.system("sysbench " + ' '.join(args) + "prepare")

    @staticmethod
    def run():
        args = Sysbench.__build_args__()
        log_path = Sysbench.__conf["log-path"]
        os.system("sysbench " + ' '.join(args) + "run" + " > " + log_path)
        return Sysbench.parse()

    @staticmethod
    def parse():
        with open(Sysbench.__conf["log-path"]) as log_file:
            lines = log_file.readlines()
            for line in lines:
                Sysbench.__match_record__(line)
            Sysbench.__match_result__(lines)
            return Sysbench.records()

    @staticmethod
    def grouped_records(n, head, tail):
        return {
            "time": TimeSeries([x[-1] for x in Sysbench.__records["time"].chunks(n)][head:tail]),
            "tps": TimeSeries([x.avg() for x in Sysbench.__records["tps"].chunks(n)][head:tail]),
            "p95_lat_ms": TimeSeries([x.max()[1] for x in Sysbench.__records["p95_lat_ms"].chunks(n)][head:tail]),
            "avg_lat_ms": TimeSeries([x.avg() for x in Sysbench.__records["avg_lat_ms"].chunks(n)][head:tail])
        }

    @staticmethod
    def eval_records(records):
        time = records["time"]
        return {
            "tps": records["tps"].eval(time),
            "p95_lat_ms": records["p95_lat_ms"].eval(time),
            "avg_lat_ms": records["avg_lat_ms"].eval(time)
        }

    @staticmethod
    def eval_records_string(records):
        eval_results = Sysbench.eval_records(records)
        return '\n'.join(
            ['\n'.join(["{}-{:20s}\t{:20s}".format(k, x, y) for x, y in v.items()]) for k, v in eval_results.items()])

    @staticmethod
    def __match_record__(line):
        obj = re.match(recordRegex, line)
        if obj is None:
            return
        time, tps, p95_lat_ms = obj.group(1, 2, 3)
        Sysbench.__records["time"].append(time)
        Sysbench.__records["tps"].append(float(tps))
        Sysbench.__records["p95_lat_ms"].append(float(p95_lat_ms))
        Sysbench.__records["avg_lat_ms"].append(1000.0 / float(tps))

    @staticmethod
    def __match_result__(lines):
        return list()

    @staticmethod
    def __build_args__():
        conf = Sysbench.__conf.copy()
        name = conf["name"]
        del conf["name"]
        del conf["log-path"]
        args = ["--{}={}".format(x[0], x[1]) for x in Sysbench.__conf.items()]
        args.append(name)
        return args
