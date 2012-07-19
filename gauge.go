package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Gauge stores a single, instantaneous Gaugable value.
// It is updated using a stored GaugeFunction.
type Gauge struct {
	value       Gaugable
	function    GaugeFunction
	lastUpdated time.Time
	lock        sync.RWMutex
}

type GaugeSnapshot struct {
	Value       Gaugable
	LastUpdated time.Time
}

type Gaugable interface {
	fmt.Stringer
}

type GaugeFunction func(*Gauge) Gaugable

func newGauge() *Gauge {
	return &Gauge{}
}

func (g *Gauge) Reset() {
	g.lock.Lock()
	g.value = nil
	g.lastUpdated = time.Time{}
	g.lock.Unlock()
}

func (g *Gauge) SetFunction(fn GaugeFunction) {
	g.lock.Lock()
	g.function = fn
	g.lock.Unlock()
}

func (g *Gauge) Update() {
	g.update(time.Now())
}

func (g *Gauge) update(now time.Time) {
	g.lock.Lock()
	if g.function != nil {
		g.value = g.function(g)
	}
	g.lastUpdated = now
	g.lock.Unlock()
}

func (g *Gauge) Snapshot() GaugeSnapshot {
	g.lock.RLock()

	r := GaugeSnapshot{
		Value:       g.value,
		LastUpdated: g.lastUpdated,
	}

	g.lock.RUnlock()
	return r
}
