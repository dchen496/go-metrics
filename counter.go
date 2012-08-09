package metrics

import (
	"sync"
)

// Counter contains a single int64, which may be incremented,
// decremented or set.
type Counter struct {
	value int64
	lock  sync.RWMutex
}

type CounterSnapshot struct {
	Value int64
}

func newCounter() *Counter {
	return &Counter{}
}

// Reset sets the Counter to zero.
func (c *Counter) Reset() {
	c.lock.Lock()
	c.value = 0
	c.lock.Unlock()
}

func (c *Counter) Inc(v int64) {
	c.lock.Lock()
	c.value += v
	c.lock.Unlock()
}

func (c *Counter) Dec(v int64) {
	c.Inc(-v)
}

func (c *Counter) Set(v int64) {
	c.lock.Lock()
	c.value = v
	c.lock.Unlock()
}

// Snapshot returns the Counter's value.
func (c *Counter) Snapshot() CounterSnapshot {
	c.lock.RLock()

	r := CounterSnapshot{
		Value: c.value,
	}

	c.lock.RUnlock()
	return r
}
