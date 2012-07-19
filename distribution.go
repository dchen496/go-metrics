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

var DistributionDefaultPercentiles = []float64{
	0.0, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 0.9999, 1.0
}

var DistributionDefaultProcessOptions = DistributionProcessOptions{
	Data:        false,
	Stats:       true,
	Percentiles: DistributionDefaultPercentiles,
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
	*Distribution
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
	d.lock.Lock()
	d.prune(time.Now())
	// not really the best way to loosen the lock
	d.lock.Unlock()
	d.lock.RLock()
	return DistributionSnapshot{d}
}

func (d *DistributionSnapshot) Unsnapshot() {
	d.lock.RUnlock()
	d.Distribution = nil
}

func (d *Distribution) Process(p Processor, name string,
	options interface{}) interface{} {

	snap := d.Snapshot()
	defer snap.Unsnapshot()

	var o *DistributionProcessOptions
	switch v := options.(type) {
	case nil:
		o = &DistributionDefaultProcessOptions
	case *MetricProcessOptions:
		o = &(v.DistributionProcessOptions)
	case *DistributionProcessOptions:
		o = v
	default:
		panic("invalid option type")
	}
	if o.Percentiles == nil {
		o.Percentiles = DistributionDefaultPercentiles
	}
	return p.ProcessDistribution(snap, name, o)
}

func (d *DistributionSnapshot) Count() uint64 {
	return d.s.Count()
}

func (d *DistributionSnapshot) Mean() float64 {
	return d.s.Mean()
}

func (d *DistributionSnapshot) Variance() float64 {
	return d.s.Variance()
}

func (d *DistributionSnapshot) StandardDeviation() float64 {
	return d.s.StandardDeviation()
}

func (d *DistributionSnapshot) Skewness() float64 {
	return d.s.Skewness()
}

func (d *DistributionSnapshot) Kurtosis() float64 {
	return d.s.Kurtosis()
}

func (d *DistributionSnapshot) Percentile(p float64) int64 {
	return d.s.Percentile(p)
}

func (d *DistributionSnapshot) Samples(limit uint64,
	begin, end *time.Time) []int64 {

	if d.Count() == 0 {
		return make([]int64, 0)
	}
	if limit > d.Count() || limit == 0 {
		limit = d.Count()
	}

	var beginNode, endNode *rbtree.Node
	var beginRank, endRank uint64

	if begin != nil {
		beginNode = d.times.LowerBound(int64(begin.Sub(d.timeBase)))
	}
	if beginNode == nil {
		beginNode = d.times.FindByRank(0)
		beginRank = 0
	} else {
		beginRank = d.times.Rank(beginNode)
	}

	if end != nil {
		endNode = d.times.UpperBound(int64(end.Sub(d.timeBase)))
	}
	if endNode == nil {
		endNode = d.times.FindByRank(d.Count() - 1)
		endRank = d.Count() - 1
	} else {
		endRank = d.times.Rank(endNode)
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
