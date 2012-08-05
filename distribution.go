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

// The Percentiles slice in a DistributionSnapshot
// contains the value of these percentiles in a Distribution. 
// 0.0 is equivalent to minimum, 0.5 is equivalent to median, 
// and 1.0 is equivalent to maximum.
var DistributionPercentiles = []float64{
	0.0, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 1.0,
}

// Distribution stores collected data samples.
// Old samples are pruned based on a specified length of time
// whenever the Distribution is written to, or before a
// DistributionSnapshot is generated.
type Distribution struct {
	s              *statistics.Sample
	times          *rbtree.Tree
	timeBase       time.Time
	window         time.Duration
	populationSize float64
	maxSampleSize  uint64
	rangeHint      [2]float64
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
	Window            time.Duration
	RangeHint         [2]float64
	LastUpdated       time.Time
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
	// d.times.Size() == d.s.Count() is always true
	return d.times.Size()
}

// SetMaxSampleSize sets how many sample elements are kept.
// The default is 1000 elements.
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

// SetWindow sets the length of time for which a element
// will be kept before a call to Prune removes it. Note that
// elements are not guaranteed to be kept if the Distribution's
// Count exceeds the limit set in SetMaxSampleSize.
// The default is 10 minutes.
func (d *Distribution) SetWindow(nsec time.Duration) {
	d.lock.Lock()
	d.window = nsec
	d.prune(time.Now())
	d.lock.Unlock()
}

// Setting a range hint for a Distribution has no effect,
// but it is exported in a DistributionSnapshot, where it may
// be used externally. For example, if set, the dashboard uses 
// this hint to size graphs.
func (d *Distribution) SetRangeHint(min, max float64) {
	d.lock.Lock()
	d.rangeHint = [2]float64{min, max}
	d.lock.Unlock()
}

// Add might insert/replace a sample into a Distribution, following 
// a random algorithm to maintain the maximum sample size set in 
// SetMaxSampleSize.
func (d *Distribution) Add(v int64) {
	d.lock.Lock()
	maxRand := int64(d.populationSize)
	if maxRand == 0 {
		d.add(v, time.Now(), 0)
	} else {
		d.add(v, time.Now(), uint64(rand.Int63n(maxRand)))
	}
	d.lock.Unlock()
}

func (d *Distribution) add(v int64, now time.Time, remove uint64) {
	d.populationSize++
	if d.size() >= d.maxSampleSize {
		if remove < d.maxSampleSize {
			n := d.times.FindByRank(remove)
			d.remove(n)
		} else {
			return
		}
	}

	se := d.s.Add(v)
	d.times.Insert(int64(now.Sub(d.timeBase)), se)
	d.prune(now)
}

// Prune removes old samples from a Distribution, according
// to the length of time set by SetWindow.
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
	d.populationSize *= float64(d.size()-1) / float64(d.size())
	d.remove(n)
}

// Snapshot returns various statistics on the Distribution.
func (d *Distribution) Snapshot() DistributionSnapshot {
	d.lock.Lock()
	d.prune(time.Now())

	var lastUpdated time.Time
	if d.size() != 0 {
		dur := time.Duration(d.times.FindByRank(d.size() - 1).Key())
		lastUpdated = d.timeBase.Add(dur)
	}

	r := DistributionSnapshot{
		Count:             d.size(),
		Mean:              d.s.Mean(),
		Variance:          d.s.Variance(),
		StandardDeviation: d.s.StandardDeviation(),
		Skewness:          d.s.Skewness(),
		Kurtosis:          d.s.Kurtosis(),
		Percentiles:       make([]int64, len(DistributionPercentiles)),
		PopulationSize:    d.populationSize,
		Window:            d.window,
		RangeHint:         d.rangeHint,
		LastUpdated:       lastUpdated,
	}
	for i, v := range DistributionPercentiles {
		r.Percentiles[i] = d.s.Percentile(v)
	}

	d.lock.Unlock()
	return r
}

// Samples returns up to limit sample elements (unlimited if limit = 0)
// from the Distribution. These are taken between a time interval
// specified with begin (inclusive) and end (non-inclusive). 
// The count is the actual number of samples in the
// time interval specified. 
//
// Special cases: 
// If begin and/or end is nil, then it is assumed to be the earliest possible // time (for begin) or the latest (for end).
// If the end time is before the begin time, an empty slice and a count of 
// -1 are returned.
// If the end time is before the Distribution was created/reset, an empty 
// slice and a count of -1 are returned.
func (d *Distribution) Samples(limit uint64,
	begin, end *time.Time) (vals []int64, count int64) {

	d.lock.RLock()
	defer d.lock.RUnlock()

	if d.size() == 0 {
		return make([]int64, 0), 0
	}
	if limit > d.size() || limit == 0 {
		limit = d.size()
	}

	var beginNode, endNode *rbtree.Node
	var beginRank, endRank uint64

	if begin != nil {
		beginNode = d.times.LowerBound(int64(begin.Sub(d.timeBase)))
		if beginNode == nil {
			return make([]int64, 0), 0
		}
		beginRank = d.times.Rank(beginNode)
	} else {
		beginRank = 0
		beginNode = d.times.FindByRank(beginRank)
	}

	if end != nil {
		if end.Before(d.timeBase) {
			return make([]int64, 0), -1
		}

		endNode = d.times.UpperBound(int64(end.Sub(d.timeBase)))
		if endNode == nil {
			return make([]int64, 0), 0
		}
		endRank = d.times.Rank(endNode)
	} else {
		endRank = d.size() - 1
		endNode = d.times.FindByRank(endRank)
	}

	if endRank < beginRank {
		return make([]int64, 0), -1
	}

	ct := endRank - beginRank + 1
	var m []int64
	if limit >= ct {
		// get everything
		m = make([]int64, ct)
		for n, i := beginNode, uint64(0); n != nil; n, i = d.times.Next(n), i+1 {
			m[i] = n.Value().(statistics.SampleElement).Value()
			if n == endNode {
				break
			}
		}
	} else {
		m = make([]int64, limit)
		s := randCombination(ct, limit)
		var i uint64 = 0
		for v := range s {
			n := d.times.FindByRank(v + beginRank)
			m[i] = n.Value().(statistics.SampleElement).Value()
			i++
		}
	}

	return m, int64(ct)
}

// Robert Floyd's sampling algorithm
// The returned map will contain (num) randomly chosen, 
// unique values chosen from [0, max).
//
// Some explanation: for n in [0, max - num), the probability
// of n not being chosen is the product of i/(i+1) as i
// ranges from max - num to max - 1, which simplifies to 
// (max - num) / max.
//
// For n in [max - num, max) the probability of not being
// chosen is 1 when i < n, (max - num)/(n + 1) when i = n, 
// and (i)/(i+1) when i > n. Multiplying all these together
// again results in (max - num) / max for the probability of
// not being included. 
func randCombination(max, num uint64) map[uint64]bool {
	s := make(map[uint64]bool)
	for i := max - num; i < max; i++ {
		// generate r in [0, i]
		r := uint64(rand.Int63n(int64(i + 1)))
		if s[r] {
			s[i] = true
		} else {
			s[r] = true
		}
	}
	return s
}
