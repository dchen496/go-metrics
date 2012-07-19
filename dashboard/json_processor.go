package dashboard

import (
	"encoding/json"
	"metrics"
	"time"
)

type JSONProcessor struct{}

func (p JSONProcessor) decodeOption(o string, m metrics.Metric) (interface{},
	error) {

	var r interface{}
	switch m.(type) {
	case *metrics.Counter:
		r = &metrics.CounterProcessOptions{}
	case *metrics.Distribution:
		r = &metrics.DistributionProcessOptions{}
	case *metrics.Gauge:
		r = &metrics.GaugeProcessOptions{}
	}
	err := json.Unmarshal([]byte(o), r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p JSONProcessor) ProcessCounter(c metrics.CounterSnapshot, name string,
	options *metrics.CounterProcessOptions) interface{} {

	m := make(map[string]interface{})
	m["name"] = name
	m["type"] = "counter"
	m["value"] = c.Value()
	m["lastUpdated"] = c.LastUpdated()

	if options.Derivatives {
		derivatives := make([]float64, 1+c.MaxDerivativeOrder())
		for i := 0; i < len(derivatives); i++ {
			derivatives[i] = c.Derivative(uint64(i), 0)
		}
		m["derivatives"] = derivatives
	}
	if options.ExpAverages {
		tc := make([]time.Duration, c.NumTimeConstants())
		exp := make([][]float64, c.NumTimeConstants())
		for i := 0; i < len(tc); i++ {
			tc[i] = c.TimeConstant(uint64(i))
			exp[i] = make([]float64, 1+c.MaxDerivativeOrder())
			for j := 0; j < len(exp[i]); j++ {
				exp[i][j] = c.Derivative(uint64(j), uint64(i+1))
			}
		}
		m["timeConstants"] = tc
		m["expAverages"] = exp
	}
	//  str, _ := json.Marshal(m)
	str, _ := json.MarshalIndent(m, "", "\t")
	return str
}

func (p JSONProcessor) ProcessDistribution(d metrics.DistributionSnapshot,
	name string, options *metrics.DistributionProcessOptions) interface{} {

	m := make(map[string]interface{})
	m["name"] = name
	m["type"] = "distribution"

	if options.Data {
		s := d.Samples(options.Limit, options.Begin, options.End)
		data := make([]int64, len(s))
		copy(data, s)
		m["data"] = data
	}

	if options.Stats {
		m["count"] = d.Count()
		m["mean"] = d.Mean()
		m["variance"] = d.Variance()
		m["standardDeviation"] = d.StandardDeviation()
		m["skewness"] = d.Skewness()
		m["kurtosis"] = d.Kurtosis()

		p := make([][2]interface{}, len(options.Percentiles))
		for i, v := range options.Percentiles {
			p[i][0] = v
			p[i][1] = d.Percentile(v)
		}
		m["percentiles"] = p
	}
	//  str, _ := json.Marshal(m)
	str, _ := json.MarshalIndent(m, "", "\t")
	return str
}

func (p JSONProcessor) ProcessGauge(g metrics.GaugeSnapshot, name string,
	options *metrics.GaugeProcessOptions) interface{} {

	m := make(map[string]interface{})
	m["name"] = name
	m["type"] = "gauge"
	if g.Value() != nil {
		m["value"] = g.Value().String()
	}
	m["lastUpdated"] = g.LastUpdated()
	//  str, _ := json.Marshal(m)
	str, _ := json.MarshalIndent(m, "", "\t")
	return str
}
