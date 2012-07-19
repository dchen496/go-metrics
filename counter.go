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
	Value         int64
	LastUpdated   time.Time
	TimeConstants []time.Duration
	Derivatives   [][]float64
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

	r := CounterSnapshot{
		Value:         c.r.Value(),
		LastUpdated:   c.r.LastUpdated(),
		TimeConstants: c.r.TimeConstants(),
		Derivatives:   c.r.Derivatives(),
	}

	c.lock.RUnlock()
	return r
}