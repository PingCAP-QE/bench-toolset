package workload

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
