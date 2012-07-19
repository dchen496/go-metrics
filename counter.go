package metrics

import (
	"metrics/statistics"
	"sync"
	"time"
)

const (
	defaultCounterDerivatives = 0
)

var defaultCounterTimeConstants = []time.Duration{}

var counterLoadAverage = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

type Counter struct {
	r    *statistics.Rate
	lock sync.RWMutex
}

type CounterSnapshot struct {
	*Counter
}

func newCounter() *Counter {
	return &Counter{
		r: statistics.NewRate(defaultCounterDerivatives,
			defaultCounterTimeConstants),
	}
}

func (c *Counter) Reset() {
	c.lock.Lock()
	c.r.Reset()
	c.lock.Unlock()
}

func (c *Counter) SetMaxDerivativeOrder(n uint64) {
	c.lock.Lock()
	c.r.SetMaxDerivativeOrder(n)
	c.lock.Unlock()
}

func (c *Counter) SetTimeConstants(tcs []time.Duration) {
	c.lock.Lock()
	c.r.SetTimeConstants(tcs)
	c.lock.Unlock()
}

func (c *Counter) Inc(v int64) {
	c.inc(v, time.Now())
}

func (c *Counter) inc(v int64, now time.Time) {
	c.lock.Lock()
	c.r.Set(c.r.Value()+v, now)
	c.lock.Unlock()
}

func (c *Counter) Dec(v int64) {
	c.Inc(-v)
}

func (c *Counter) Set(v int64) {
	c.set(v, time.Now())
}

func (c *Counter) set(v int64, now time.Time) {
	c.lock.Lock()
	c.r.Set(v, now)
	c.lock.Unlock()
}

func (c *Counter) Snapshot() CounterSnapshot {
	c.lock.RLock()
	return CounterSnapshot{c}
}

func (c *CounterSnapshot) Unsnapshot() {
	c.lock.RUnlock()
	c.Counter = nil
}

var CounterDefaultProcessOptions = CounterProcessOptions{
	Derivatives: true,
}

func (c *Counter) Process(p Processor, name string,
	options interface{}) interface{} {

	snap := c.Snapshot()
	// don't make a mess if the something panics
	defer snap.Unsnapshot()

	var o *CounterProcessOptions
	switch v := options.(type) {
	case nil:
		o = &CounterDefaultProcessOptions
	case *MetricProcessOptions:
		o = &(v.CounterProcessOptions)
	case *CounterProcessOptions:
		o = v
	default:
		panic("invalid option type")
	}
	return p.ProcessCounter(snap, name, o)
}

func (c *CounterSnapshot) Value() int64 {
	return c.r.Value()
}

func (c *CounterSnapshot) LastUpdated() time.Time {
	return c.r.LastUpdated()
}

func (c *CounterSnapshot) NumTimeConstants() uint64 {
	return c.r.NumTimeConstants()
}

func (c *CounterSnapshot) TimeConstant(index uint64) time.Duration {
	return c.r.TimeConstant(index)
}

func (c *CounterSnapshot) MaxDerivativeOrder() uint64 {
	return c.r.MaxDerivativeOrder()
}

func (c *CounterSnapshot) Derivative(order uint64,
	timeConstantIndex uint64) float64 {

	return c.r.Derivative(order, timeConstantIndex)
}
