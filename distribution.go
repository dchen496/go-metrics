package metrics

import (
	"math/rand"
	"metrics/rbtree"
	"metrics/statistics"
	"sync"
	"time"
)

const (
	distributionDefaultMaxSamples = 1000
	distributionDefaultWindow     = time.Minute * 10
)

var DistributionPercentiles = []float64{
	0.0, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 0.9999, 1.0,
}

// Distribution stores collected data samples.
// Old samples are pruned based on a specified length of time
// whenever the Distribution is written to, or before a
// DistributionValue is generated.
// Distribution also computes certain statistics on the fly for
// fast retrieval.
type Distribution struct {
	s              *statistics.Sample
	times          *rbtree.Tree
	timeBase       time.Time
	window         time.Duration
	populationSize float64
	maxSampleSize  uint64
	lock           sync.RWMutex
}

type DistributionSnapshot struct {
	Count             uint64
	Mean              float64
	Variance          float64
	StandardDeviation float64
	Skewness          float64
	Kurtosis          float64
	Percentiles       []int64
	PopulationSize    float64
}

func newDistribution() *Distribution {
	return &Distribution{
		s:             statistics.NewSample(),
		times:         rbtree.New(),
		timeBase:      time.Now(),
		window:        distributionDefaultWindow,
		maxSampleSize: distributionDefaultMaxSamples,
	}
}

// Reset deletes all samples and statistics from a Distribution.
func (d *Distribution) Reset() {
	d.lock.Lock()
	d.s = statistics.NewSample()
	d.populationSize = 0
	d.times = rbtree.New()
	d.lock.Unlock()
}

func (d *Distribution) size() uint64 {
	// d.times.Size() == d.s.Count() 
	return d.times.Size()
}

func (d *Distribution) SetMaxSampleSize(n uint64) {
	d.lock.Lock()
	d.maxSampleSize = n
	for d.size() > d.maxSampleSize {
		r := rand.Int63n(int64(d.size()))
		node := d.times.FindByRank(uint64(r))
		d.removeFromPopulation(node)
	}
	d.lock.Unlock()
}

func (d *Distribution) SetWindow(nsec time.Duration) {
	d.lock.Lock()
	d.window = nsec
	d.prune(time.Now())
	d.lock.Unlock()
}

// Add inserts a sample into a Distribution.
func (d *Distribution) Add(v int64) {
	maxRand := int64(d.populationSize)
	if maxRand == 0 {
		d.add(v, time.Now(), 0)
	} else {
		d.add(v, time.Now(), uint64(rand.Int63n(maxRand)))
	}
}

func (d *Distribution) add(v int64, now time.Time, remove uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.populationSize++
	if d.s.Count() >= d.maxSampleSize {
		r := remove
		if r < d.maxSampleSize {
			n := d.times.FindByRank(r)
			d.remove(n)
		} else {
			return
		}
	}

	se := d.s.Add(v)
	d.times.Insert(int64(now.Sub(d.timeBase)), se)
	d.prune(now)
}

func (d *Distribution) Prune() {
	d.lock.Lock()
	d.prune(time.Now())
	d.lock.Unlock()
}

func (d *Distribution) prune(now time.Time) {
	if d.window == 0 {
		return
	}

	lastKey := int64(now.Add(-d.window).Sub(d.timeBase))
	for tn := d.times.UpperBound(lastKey); tn != nil; {
		prev := tn
		tn = d.times.Prev(tn)
		d.removeFromPopulation(prev)
	}
}

func (d *Distribution) remove(n *rbtree.Node) {
	d.s.Remove(n.Value().(statistics.SampleElement))
	d.times.RemoveNode(n)
}

func (d *Distribution) removeFromPopulation(n *rbtree.Node) {
	d.populationSize *= float64(d.s.Count()-1) / float64(d.s.Count())
	d.remove(n)
}

func (d *Distribution) Snapshot() DistributionSnapshot {
	d.Prune()

	d.lock.RLock()

	r := DistributionSnapshot{
		Count:             d.s.Count(),
		Mean:              d.s.Mean(),
		Variance:          d.s.Variance(),
		StandardDeviation: d.s.StandardDeviation(),
		Skewness:          d.s.Skewness(),
		Kurtosis:          d.s.Kurtosis(),
		Percentiles:       make([]int64, len(DistributionPercentiles)),
		PopulationSize:    d.populationSize,
	}
	for i, v := range DistributionPercentiles {
		r.Percentiles[i] = d.s.Percentile(v)
	}

	d.lock.RUnlock()
	return r
}

func (d *Distribution) Samples(limit uint64,
	begin, end *time.Time) []int64 {

	d.lock.RLock()
	defer d.lock.RUnlock()

	if d.size() == 0 {
		return make([]int64, 0)
	}
	if limit > d.size() || limit == 0 {
		limit = d.size()
	}

	var beginNode, endNode *rbtree.Node
	var beginRank, endRank uint64

	if begin != nil {
		beginNode = d.times.LowerBound(int64(begin.Sub(d.timeBase)))
	}
	if beginNode != nil {
		beginRank = d.times.Rank(beginNode)
	} else {
		beginRank = 0
		beginNode = d.times.FindByRank(beginRank)
	}

	if end != nil {
		endNode = d.times.UpperBound(int64(end.Sub(d.timeBase)))
	}
	if endNode != nil {
		endRank = d.times.Rank(endNode)
	} else {
		endRank = d.size() - 1
		endNode = d.times.FindByRank(endRank)
	}

	if endRank < beginRank {
		return make([]int64, 0)
	}

	diff := endRank - beginRank
	var m []int64
	if limit >= diff {
		// get everything
		m = make([]int64, diff+1)
		for n, i := beginNode, uint64(0); n != nil; n, i = d.times.Next(n), i+1 {
			m[i] = n.Value().(statistics.SampleElement).Value()
			if n == endNode {
				break
			}
		}
	} else {
		m = make([]int64, limit)
		s := randCombination(diff+1, limit)
		var i uint64 = 0
		for v := range s {
			n := d.times.FindByRank(v + beginRank)
			m[i] = n.Value().(statistics.SampleElement).Value()
			i++
		}
	}
	return m
}

// Robert Floyd's sampling algorithm
// s will contain (limit) randomly chosen, unique values
// chosen from [0, max)
func randCombination(max, num uint64) map[uint64]bool {
	s := make(map[uint64]bool)
	for i := max - num; i < max; i++ {
		r := uint64(rand.Int63n(int64(i + 1)))
		if s[r] {
			s[i] = true
		} else {
			s[r] = true
		}
	}
	return s
}
