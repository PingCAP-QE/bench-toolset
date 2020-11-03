package bench

import "time"

type Workload interface {
	Start() error
}

type Sysbench struct {
	Host string
	User string
	Port uint64
	Db   string

	Tables    uint64
	TableSize uint64

	Name           string
	Threads        uint64
	Time           time.Duration
	ReportInterval time.Duration
}

func (s *Sysbench) Start() error {
	return nil
}

type Ycsb struct {
	Type     string
	Workload string

	Host string
	User string
	Port uint64
	Db   string

	Threads    uint64
	Operations uint64
	Records    uint64
	Fields     uint64
}

type Tpcc struct {
}
