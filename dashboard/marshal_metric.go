package dashboard

import (
	"fmt"
	"metrics"
	"time"
)

type TypeValue struct {
	Type  string
	Value interface{}
}

func TypeValueMetric(me metrics.Metric) TypeValue {
	var tv TypeValue

	switch m := me.(type) {
	case *metrics.Counter:
		tv.Type = "counter"
		tv.Value = m.Snapshot()

	case *metrics.Distribution:
		tv.Type = "distribution"
		tv.Value = m.Snapshot()

	case *metrics.Gauge:
		snapshot := m.Snapshot()
		var stringified struct {
			Value       string
			LastUpdated time.Time
		}
		if snapshot.Value != nil {
			stringified.Value = snapshot.Value.String()
		}
		stringified.LastUpdated = snapshot.LastUpdated

		tv.Type = "gauge"
		tv.Value = stringified

	case *metrics.Meter:
		tv.Type = "meter"
		tv.Value = m.Snapshot()
	}

	return tv
}

func TypeValueSamples(d *metrics.Distribution,
	beginstr, endstr, limitstr string) TypeValue {

	var tv TypeValue
	tv.Type = "distribution_sample"

	var begin, end time.Time
	beginptr, endptr := &begin, &end
	begin, err := time.Parse(time.RFC3339, beginstr)
	if err != nil {
		beginptr = nil
	}
	end, err = time.Parse(time.RFC3339, endstr)
	if err != nil {
		endptr = nil
	}

	var limit uint64
	fmt.Sscanf(limitstr, "%d", &limit)

	var t struct {
		Samples []int64
		Count   uint64
	}
	t.Samples, t.Count = d.Samples(limit, beginptr, endptr)

	tv.Value = t

	return tv
}
