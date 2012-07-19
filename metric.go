package metrics

type Metric interface {
	Reset()
	Process(p Processor, name string, options interface{}) interface{}
}
