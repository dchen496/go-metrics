package metrics

import (
	"sync"
	"time"
)

// Counter contains a single int64, which may be incremented,
// decremented or set. 
type Counter struct {
	value       int64
	lastUpdated time.Time
	lock        sync.RWMutex
}

type CounterSnapshot struct {
	Value       int64
	LastUpdated time.Time
}

func newCounter() *Counter {
	return &Counter{}
}

// Reset sets the Counter to zero and sets the last updated
// time to the zero time.
func (c *Counter) Reset() {
	c.lock.Lock()
	c.value = 0
	c.lastUpdated = time.Time{}
	c.lock.Unlock()
}

func (c *Counter) Inc(v int64) {
	c.lock.Lock()
	c.value += v
	c.lastUpdated = time.Now()
	c.lock.Unlock()
}

func (c *Counter) Dec(v int64) {
	c.Inc(-v)
}

func (c *Counter) Set(v int64) {
	c.lock.Lock()
	c.value = v
	c.lastUpdated = time.Now()
	c.lock.Unlock()
}

// Snapshot returns the Counter's value.
func (c *Counter) Snapshot() CounterSnapshot {
	c.lock.RLock()

	r := CounterSnapshot{
		Value:       c.value,
		LastUpdated: c.lastUpdated,
	}

	c.lock.RUnlock()
	return r
}
