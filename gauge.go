package metrics

import (
	"fmt"
	"sync"
)

// Gauge stores a single, instantaneous Gaugable value.
// It is updated using a stored GaugeFunction.
type Gauge struct {
	value    Gaugable
	function GaugeFunction
	lock     sync.RWMutex
}

type GaugeSnapshot struct {
	Value Gaugable
}

type Gaugable interface {
	fmt.Stringer
}

// A gauge's GaugeFunction is called each time it is updated.
type GaugeFunction func() Gaugable

func newGauge() *Gauge {
	return &Gauge{}
}

// Reset clears a Gauge's value.
func (g *Gauge) Reset() {
	g.lock.Lock()
	g.value = nil
	g.lock.Unlock()
}

// SetFunction associates a GaugeFunction to a Gauge.
func (g *Gauge) SetFunction(fn GaugeFunction) {
	g.lock.Lock()
	g.function = fn
	g.lock.Unlock()
}

// Update calls the function set in SetFunction, passing
// the Gauge as its first argument. It stores the function's
// return value as the Gauge's value.
func (g *Gauge) Update() {
	g.lock.Lock()
	if g.function != nil {
		g.value = g.function()
	}
	g.lock.Unlock()
}

// Snapshot returns the value of a Gauge.
// The value is the Gaugable object. If a string is
// required, it must be converted manually using the
// Gaugable's String() method.
func (g *Gauge) Snapshot() GaugeSnapshot {
	g.lock.RLock()

	r := GaugeSnapshot{
		Value: g.value,
	}

	g.lock.RUnlock()
	return r
}
