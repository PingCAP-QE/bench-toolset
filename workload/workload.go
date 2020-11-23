package workload

type Record struct {
	Type    string
	Count   float64
	AvgLatInMs float64
	P99LatInMs float64
}

type Workload interface {
	Prepare() error
	Start() error
	Records() ([]*Record, error)
}
