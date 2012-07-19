package metrics

import ()

type testProcessor struct{}

func (t testProcessor) ProcessCounter(c CounterSnapshot, name string,
	options *CounterProcessOptions) interface{} {
	if name != "test" {
		return -1
	}
	if c.Value() != 1357 {
		return -2
	}
	if options.Derivatives != true || options.ExpAverages != false {
		return -3
	}
	if c.LastUpdated() != testTime {
		return -4
	}
	return 1
}

func (t testProcessor) ProcessDistribution(d DistributionSnapshot, name string,
	options *DistributionProcessOptions) interface{} {
	if name != "test" {
		return -1
	}
	if d.Count() != 4 {
		return -2
	}
	if options.Data != true || options.Limit != 1234 || options.Stats != false {
		return -3
	}
	return 2
}

func (t testProcessor) ProcessGauge(g GaugeSnapshot, name string,
	options *GaugeProcessOptions) interface{} {
	if name != "test" {
		return -1
	}
	v := testGaugable{value: 5, status: true}
	if *(g.Value().(*testGaugable)) != v {
		return -2
	}
	if g.LastUpdated() != testTime {
		return -4
	}
	return 3
}
