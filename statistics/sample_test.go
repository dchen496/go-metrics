package statistics

import (
	"fmt"
	"testing"
)

const (
	sampleToAdd    = 1654918
	sampleToRemove = -4176556
)

func testSampleInit() *Sample {
	s := NewSample()
	m := make([]SampleElement, len(testSampleSet))
	for i, v := range testSampleSet {
		m[i] = s.Add(v)
	}
	return s
}

func testSampleAdd(s *Sample) {
	s.Add(sampleToAdd)
}

func testSampleRemove(s *Sample) {
	s.Remove(SampleElement{s.values.Find(sampleToRemove)})
}

func testWithChanges(t *testing.T,
	s *Sample, name string, testfunc func(*Sample) interface{},
	original interface{}, added interface{}, removed interface{}) {

	testCompare(t, name, testfunc(s), original)
	testSampleAdd(s)
	testCompare(t, name, testfunc(s), added)
	testSampleRemove(s)
	testCompare(t, name, testfunc(s), removed)
}

func TestSampleCount(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return int(s.Count()) }
	testWithChanges(t, s, "sample Count", testfunc,
		len(testSampleSet), len(testSampleSet)+1, len(testSampleSet))
}

func TestSampleMean(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return s.Mean() }
	testWithChanges(t, s, "sample Mean", testfunc,
		311618.12, 337957.3333333333, 428247.6)
}

var testSamplePercentiles = [][]float64{
	{0.0, -9988298, -9988298, -9988298},
	{1.0, 9694132, 9694132, 9694132},
	{0.5, 647511, 647511, 745446},
	{0.75, 5006064, 5006064, 5006064},
	{0.25, -4176556, -3796327, -3796327},
	{0.99, 9694132, 9694132, 9694132},
}

func TestSamplePercentile(t *testing.T) {
	for _, v := range testSamplePercentiles {
		s := testSampleInit()
		testfunc := func(s *Sample) interface{} {
			return float64(s.Percentile(v[0]))
		}
		testWithChanges(t, s, fmt.Sprint("percentile ", v[0]), testfunc,
			v[1], v[2], v[3])
	}
}

func TestSampleVariance(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return s.Variance() }
	testWithChanges(t, s, "sample Variance", testfunc,
		3.452546851438476e13, 3.387034060620703e13, 3.413731802164837e13)
}

func TestSampleStandardDeviation(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return s.StandardDeviation() }
	testWithChanges(t, s, "sample StandardDeviation", testfunc,
		5.875837686184393e6, 5.819823073445363e6, 5.842714952969070e6)
}

func TestSampleSkewness(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return s.Skewness() }
	testWithChanges(t, s, "sample Skewness", testfunc,
		-0.004110084618925392, -0.01760862112253390, -0.05508609710752556)
}

func TestSampleKurtosis(t *testing.T) {
	s := testSampleInit()
	testfunc := func(s *Sample) interface{} { return s.Kurtosis() }
	testWithChanges(t, s, "sample Kurtosis", testfunc,
		1.943071395965790, 1.978048762493122, 1.982319611215673)
}

func TestSampleRemoveLastTwo(t *testing.T) {
	s := NewSample()
	a := s.Add(1)
	b := s.Add(2)
	s.Remove(a)
	if !(s.mean == 2 && s.secondCMtimesN == 0 &&
		s.thirdCMtimesN == 0 && s.fourthCMtimesN == 0 && s.Count() == 1) {
		t.Errorf("Sample stats not reset. "+
			"Got: mean:%f 2CM_N:%f 3CM_N:%f 4CM_N:%f count %d",
			s.mean, s.secondCMtimesN, s.thirdCMtimesN, s.fourthCMtimesN, s.Count())
	}
	s.Remove(b)
	if !(s.mean == 0 && s.secondCMtimesN == 0 &&
		s.thirdCMtimesN == 0 && s.fourthCMtimesN == 0 && s.Count() == 0) {
		t.Errorf("Sample stats not reset. "+
			"Got: mean:%f 2CM_N:%f 3CM_N:%f 4CM_N:%f count %d",
			s.mean, s.secondCMtimesN, s.thirdCMtimesN, s.fourthCMtimesN, s.Count())
	}
}
