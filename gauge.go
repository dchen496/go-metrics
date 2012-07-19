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
	*Gauge
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
	return GaugeSnapshot{g}
}

func (g *GaugeSnapshot) Unsnapshot() {
	g.lock.RUnlock()
	g.Gauge = nil
}

var GaugeDefaultProcessOptions = GaugeProcessOptions{}

func (g *Gauge) Process(p Processor, name string,
	options interface{}) interface{} {

	snap := g.Snapshot()
	defer snap.Unsnapshot()

	var o *GaugeProcessOptions
	switch v := options.(type) {
	case nil:
		o = &GaugeDefaultProcessOptions
	case *MetricProcessOptions:
		o = &(v.GaugeProcessOptions)
	case *GaugeProcessOptions:
		o = v
	default:
		panic("invalid option type")
	}
	return p.ProcessGauge(snap, name, o)
}

func (g *GaugeSnapshot) Value() Gaugable {
	return g.value
}

func (g *GaugeSnapshot) LastUpdated() time.Time {
	return g.lastUpdated
}
