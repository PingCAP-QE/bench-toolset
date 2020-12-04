package workload

type Record struct {
	Type       string
	Tag        string
	Count      float64
	AvgLatInMs float64
	// metric name -> value
	Payload map[string]interface{}
}

type Workload interface {
	Prepare() error
	Start() error
	Records() ([]*Record, []*Record, error)
}
