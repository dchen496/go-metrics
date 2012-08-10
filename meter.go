package metrics

import (
	"metrics/statistics"
	"sync"
	"time"
)

const (
	defaultMeterDerivatives = 1
)

var defaultMeterTimeConstants = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

// Meter stores a int64 value like a Counter.
// Unlike a Counter, a Meter also calculates instantaneous rate of
// change, as well as 1-min, 5-min and 15-min exponentially weighted
// averages of the value and the rate of change.
type Meter struct {
	r    *statistics.Rate
	lock sync.RWMutex
}

type MeterSnapshot struct {
	Value         int64
	LastUpdated   time.Time
	TimeConstants []time.Duration
	Derivatives   [][]float64
}

func newMeter() *Meter {
	return &Meter{
		r: statistics.NewRate(defaultMeterDerivatives,
			defaultMeterTimeConstants),
	}
}

// Reset clears a Meter's values.
func (m *Meter) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.r.Reset()
}

func (m *Meter) Inc(v int64) {
	m.inc(v, time.Now())
}

func (m *Meter) inc(v int64, now time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.r.Set(m.r.Value()+v, now)
}

func (m *Meter) Dec(v int64) {
	m.Inc(-v)
}

func (m *Meter) Set(v int64) {
	m.set(v, time.Now())
}

func (m *Meter) set(v int64, now time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.r.Set(v, now)
}

// Snapshot returns the Meter's value along with
// the rate of change and exponentionally weighted averages.
// The Derivatives array is first indexed by derivative order
// (0 = value, 1 = rate of change), then by the length of time
// for the average (0 = instantaneous, 1 = 1 minute,
// 2 = 5 minutes, 3 = 15 minutes).
func (m *Meter) Snapshot() MeterSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()

	r := MeterSnapshot{
		Value:       m.r.Value(),
		LastUpdated: m.r.LastUpdated(),
		Derivatives: m.r.Derivatives(),
	}

	return r
}
