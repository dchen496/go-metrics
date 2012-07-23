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

func (m *Meter) Reset() {
	m.lock.Lock()
	m.r.Reset()
	m.lock.Unlock()
}

func (m *Meter) Inc(v int64) {
	m.inc(v, time.Now())
}

func (m *Meter) inc(v int64, now time.Time) {
	m.lock.Lock()
	m.r.Set(m.r.Value()+v, now)
	m.lock.Unlock()
}

func (m *Meter) Dec(v int64) {
	m.Inc(-v)
}

func (m *Meter) Set(v int64) {
	m.set(v, time.Now())
}

func (m *Meter) set(v int64, now time.Time) {
	m.lock.Lock()
	m.r.Set(v, now)
	m.lock.Unlock()
}

func (m *Meter) Snapshot() MeterSnapshot {
	m.lock.RLock()

	r := MeterSnapshot{
		Value:       m.r.Value(),
		LastUpdated: m.r.LastUpdated(),
		Derivatives: m.r.Derivatives(),
	}

	m.lock.RUnlock()
	return r
}
