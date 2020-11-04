package bench

type Benchmark interface {
	Run() error
	Report() (string, error)
}
