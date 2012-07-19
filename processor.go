package metrics

import (
	"time"
)

type Processor interface {
	ProcessCounter(c CounterSnapshot, name string,
		options *CounterProcessOptions) interface{}

	ProcessDistribution(d DistributionSnapshot, name string,
		options *DistributionProcessOptions) interface{}

	ProcessGauge(g GaugeSnapshot, name string,
		options *GaugeProcessOptions) interface{}
}

type CounterProcessOptions struct {
	Derivatives bool
	ExpAverages bool
}

type DistributionProcessOptions struct {
	Data  bool
	Limit uint64
	Begin *time.Time
	End   *time.Time

	Stats       bool
	Percentiles []float64
}

type GaugeProcessOptions struct{}

type MetricProcessOptions struct {
	CounterProcessOptions
	DistributionProcessOptions
	GaugeProcessOptions
}
