from prometheus_api_client import PrometheusConnect

TidbQpsQuery = "sum(rate(tidb_executor_statement_total[1m]))"
TidbQpsQueryByType = "sum(rate(tidb_executor_statement_total[1m])) by (type)"


class Connection:
    def __init__(self, host):
        self.prom = PrometheusConnect(host)

    def query(self, query, start, end):
        return self.prom.custom_query_range(query, start, end, "15s")

    def tidb_qps(self, start, end):
        return self.prom.custom_query_range(TidbQpsQuery, start, end, "15s")

    def tidb_qps_by_type(self, start, end):
        return self.prom.custom_query_range(TidbQpsQueryByType, start, end, "15s")
