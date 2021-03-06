package statistics

import (
	"fmt"
	"testing"
	"time"
)

func testRateInit() *Rate {
	r := NewRate(2,
		[]time.Duration{
			time.Minute,
			5 * time.Minute,
			15 * time.Minute,
		},
	)

	timeBase := time.Time{}
	for ind, v := range testSampleSet {
		r.Set(v, timeBase.Add(testTimeOffsets[ind]))
	}
	return r
}

func TestRateValue(t *testing.T) {
	r := testRateInit()
	testCompare(t, "rate Value", r.Value(), int64(7384336))
}

func TestRateLastUpdated(t *testing.T) {
	r := testRateInit()
	timeBase := time.Time{}
	testCompare(t, "rate LastUpdated", r.LastUpdated(), timeBase.Add(98853))
}

var testRateDerivatives = [][]float64{
	{7.384336e+06, -1.8973316454608273,
		-0.37946685829958715, -0.12648898023846042},
	{2.2635146443514645e+11, -38178.43996419731,
		-7635.727284335569, -2545.2429415487986},
	{3.728368586936403e+17, 3.770026942133821e+09,
		7.541784554752643e+08, 2.5135351986701208e+08},
}

func TestRateDerivatives(t *testing.T) {
	r := testRateInit()
	d := r.Derivatives()
	for i, order := range d {
		for j, val := range order {
			testCompare(t,
				fmt.Sprintf("rate Derivatives order %d time constant %d", i, j),
				val, testRateDerivatives[i][j],
			)
		}
	}
}

func TestSetDerivativeOrder(t *testing.T) {
	//TODO
}

func TestSetTimeConstants(t *testing.T) {
	//TODO
}
