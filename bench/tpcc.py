import re
import datetime

recordRegexp = re.compile(
    r'\[Current\]\s([\w]+)\s-\sTakes\(s\):\s([\d\.]+),\sCount:\s(\d+),\sTPM:\s([\d\.]+),\sSum\(ms\):\s([\d\.]+),\sAvg\(ms\):\s([\d\.]+),\s50th\(ms\):\s([\d\.]+),\s90th\(ms\):\s([\d\.]+),\s95th\(ms\):\s([\d\.]+),\s99th\(ms\):\s([\d\.]+),\s99\.9th\(ms\):\s([\d\.]+)')
summaryRegexp = re.compile(
    r'\[Summary\]\s([\w]+)\s-\sTakes\(s\):\s([\d\.]+),\sCount:\s(\d+),\sTPM:\s([\d\.]+),\sSum\(ms\):\s([\d\.]+),\sAvg\(ms\):\s([\d\.]+),\s50th\(ms\):\s([\d\.]+),\s90th\(ms\):\s([\d\.]+),\s95th\(ms\):\s([\d\.]+),\s99th\(ms\):\s([\d\.]+),\s99\.9th\(ms\):\s([\d\.]+)')


class Tpcc:
    __conf = {
        "host": "",
        "user": "",
        "port": 0,
        "db": "",
        "warehouses": 0,
        "threads": 0,
        "time": datetime.datetime,
        "log_path": "",
    }
    __records = []

    @staticmethod
    def config(name):
        return Tpcc.__conf[name]

    @staticmethod
    def set(name, value):
        if name in Tpcc.__conf.keys():
            Tpcc.__conf[name] = value
        else:
            raise NameError("config item not accepted in tpcc")

    @staticmethod
    def records():
        return None

    @staticmethod
    def prepare():
        return None

    @staticmethod
    def run():
        return None

    @staticmethod
    def parse():
        return None

    @staticmethod
    def __build_args__():
        return []
