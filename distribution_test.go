package metrics

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func testCompareSlices(a []int64, b []int64) bool {
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func testDistributionInit() *Distribution {
	d := newDistribution()
	d.SetWindow(time.Hour * 24 * 365 * 200)
	d.add(12, testTime, 0)
	d.add(-9, testTime.Add(1), 0)
	d.add(30, testTime.Add(1), 0)
	d.add(12, testTime.Add(2), 0)
	return d
}

func TestDistributionAdd(t *testing.T) {
	d := testDistributionInit()
	d.add(9, testTime.Add(3), 0)

	if d.times.Size() != 5 {
		t.Errorf("Number not added to times tree, got count = %d, expected %d",
			d.times.Size(), 4)
	}
	if d.size() != 5 {
		t.Errorf("Number not added to Sample, got count = %d, expected %d",
			d.size(), 4)
	}

	s := d.Samples(0, nil, nil)
	expected := []int64{12, -9, 30, 12, 9}
	if !testCompareSlices(s, expected) {
		t.Errorf("Wrong sample slice, got %v expected %v", s, expected)
	}
}

func TestDistributionProbabilisticAdd(t *testing.T) {
	skip := 0
	for i := uint64(0); i < 7; i++ {
		d := testDistributionInit()
		d.SetMaxSampleSize(4)
		d.add(1, testTime.Add(3), 4)
		d.add(1, testTime.Add(3), 4)
		d.add(1, testTime.Add(3), 4)

		if d.populationSize != 7 {
			t.Errorf("Wrong population size, got %f expected %d",
				d.populationSize, 7)
		}

		d.add(9, testTime.Add(3), i)
		s := d.Samples(0, nil, nil)
		expected := []int64{12, -9, 30, 12}
		if !testCompareSlices(s, expected) {
			skip++
		}

	}
	if skip != 4 {
		t.Errorf("Wrong number of skipped adds, got %d expected %d", skip, 4)
	}
}

func TestDistributionSetMaxSampleSize(t *testing.T) {
	d := testDistributionInit()
	d.SetMaxSampleSize(50)
	for i := int64(0); i < 100; i++ {
		d.add(i, testTime.Add(3), 0)
	}
	if d.size() != 50 {
		t.Errorf("Wrong number of samples after setting maximum, "+
			"got %d expected %d", d.size(), 50)
	}
	if math.Abs(d.populationSize-104.0) > 1e-13 {
		t.Errorf("Wrong population size after setting maximum, "+
			"got %f expected %f", d.populationSize, 104.0)
	}

	d.SetMaxSampleSize(25)
	if d.size() != 25 {
		t.Errorf("Wrong number of samples after setting maximum,"+
			" got %d expected %d", d.size(), 25)
	}
	if math.Abs(d.populationSize-52.0) > 1e-13 {
		t.Errorf("Wrong population size after setting maximum, "+
			"got %f expected %f", d.populationSize, 52.0)
	}
}

func TestDistributionPrune(t *testing.T) {
	d := testDistributionInit()

	d.prune(testTime.Add(0).Add(d.window))
	if d.size() != 4 {
		t.Errorf("Wrong size after fake pruning, got %d expected %d",
			d.size(), 4)
	}

	d.prune(testTime.Add(1).Add(d.window))
	if d.size() != 3 {
		t.Errorf("Wrong size after first pruning, got %d expected %d",
			d.size(), 3)
	}

	d.prune(testTime.Add(1).Add(d.window))
	if d.size() != 3 {
		t.Errorf("Wrong size after repeated first pruning, got %d expected %d",
			d.size(), 3)
	}

	d.prune(testTime.Add(3).Add(d.window))
	if d.size() != 0 {
		t.Errorf("Wrong size after second pruning, got %d expected %d",
			d.size(), 0)
	}

	d = testDistributionInit()
	d.window = 1
	d.add(3, testTime.Add(3), 0)
	if d.size() != 2 {
		t.Errorf("Wrong size after prune during add, got %d expected %d",
			d.size(), 2)
	}
}

func TestDistributionRemoveFromPopulation(t *testing.T) {
	d := testDistributionInit()
	d.removeFromPopulation(d.times.FindByRank(2))

	s := d.Samples(0, nil, nil)
	expected := []int64{12, -9, 12}
	if !testCompareSlices(s, expected) {
		t.Errorf("Wrong samples after remove, got %v expected %v", s, expected)
	}

	if d.populationSize != 3 {
		t.Errorf("Wrong population size after remove, got %d expected %d",
			d.populationSize, 3)
	}
}

func TestDistributionSamples(t *testing.T) {
	d := testDistributionInit()
	d.Reset()
	baseExpected := make([]int64, 100)
	for i := int64(0); i < 100; i++ {
		d.add(i, testTime.Add(time.Duration(i)), 0)
		baseExpected[i] = i
	}

	s := d.Samples(0, nil, nil)
	if !testCompareSlices(s, baseExpected) {
		t.Errorf("Samples (all) returned wrong slice, got %v expected %v",
			s, baseExpected)
	}

	beginTime := testTime.Add(25)
	endTime := testTime.Add(75)

	s = d.Samples(0, &beginTime, nil)
	expected := baseExpected[25:]
	if !testCompareSlices(s, expected) {
		t.Errorf("Samples (begin) returned wrong slice, got %v expected %v",
			s, expected)
	}

	s = d.Samples(0, nil, &endTime)
	expected = baseExpected[:75]
	if !testCompareSlices(s, expected) {
		t.Errorf("Samples (end) returned wrong slice, got %v expected %v",
			s, expected)
	}

	s = d.Samples(0, &beginTime, &endTime)
	expected = baseExpected[25:75]
	if !testCompareSlices(s, expected) {
		t.Errorf("Samples (begin, end) returned wrong slice, got %v expected %v",
			s, expected)
	}

	s = d.Samples(30, &beginTime, &endTime)
	if len(s) != 30 {
		t.Errorf("Samples (limit) returned wrong slice size, got %d expected %d",
			len(s), 30)
	}
	m := make(map[int64]bool)
	for _, v := range s {
		if m[v] {
			t.Errorf("Samples (limit) returned slice with duplicate %d", v)
		}
		m[v] = true
	}
}

func BenchmarkDistributionAdd(b *testing.B) {
	d := newDistribution()
	for i := 0; i < b.N; i++ {
		d.Add(rand.Int63())
		if i%10000 == 0 {
			d.Reset()
		}
	}
}