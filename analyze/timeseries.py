import operator
import statistics


class TimeSeries:
    values = list()

    def __init__(self, values):
        self.values = values

    def min(self):
        return min(enumerate(self.values), key=operator.itemgetter(1))

    def max(self):
        return max(enumerate(self.values), key=operator.itemgetter(1))

    def stdev(self):
        return statistics.stdev(self.values)

    def avg(self):
        return sum(self.values) / len(self.values)

    def append(self, n):
        self.values.append(n)

    def chunks(self, n):
        return filter(lambda l: len(l.values) == n,
                      [TimeSeries(self.values[i:i + n]) for i in range(0, len(self.values), n)])

    def eval(self, tags):
        jitter = self.stdev() / self.avg()
        max_index, max_value = self.max()
        min_index, min_value = self.min()
        positive_max = max_value / self.avg() - 1
        negative_max = min_value / self.avg() - 1
        result = {"avg": "{:.2f}".format(self.avg()), "jitter": "{:.2f}%".format(jitter * 100)}
        if tags is None:
            result["positive-max"] = "{:.2f}%".format(positive_max * 100)
            result["negative-max"] = "{:.2f}%".format(negative_max * 100)
        else:
            result["positive-max"] = "{:.2f}% at {}".format(positive_max * 100, tags[max_index])
            result["negative-max"] = "{:.2f}% at {}".format(negative_max * 100, tags[min_index])
        return result

    def __getitem__(self, key):
        if isinstance(key, slice):
            return self.values[key.start: key.step]
        elif isinstance(key, int):
            return self.values[key]

    def __repr__(self):
        return "{}".format(self.values)

    def __len__(self):
        return len(self.values)
