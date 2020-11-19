package bench

type Benchmark interface {
	Prepare() error
	Run() error
	Results() ([]*Result, error)
}

type Result struct {
	Type  string
	Name  string
	Value string
}
