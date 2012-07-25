package metrics

import (
	"fmt"
	"reflect"
	"sync"
)

// Registry stores metrics and their unique names.
type Registry struct {
	name    string
	metrics map[string]Metric
	lock    sync.RWMutex
}

func NewRegistry(name string) *Registry {
	return &Registry{
		name:    name,
		metrics: make(map[string]Metric),
	}
}

func realType(tyep interface{}) string {
	t := reflect.TypeOf(tyep)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

func (r *Registry) register(tyep interface{}, name string, m Metric) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	fullName := fmt.Sprintf("%s.%s", realType(tyep), name)

	if _, exists := r.metrics[fullName]; exists {
		return false
	}
	r.metrics[fullName] = m
	return true
}

// NewCounter creates a counter and registers it with the receiver.
func (r *Registry) NewCounter(tyep interface{}, name string) *Counter {
	m := newCounter()
	if r.register(tyep, name, m) {
		return m
	}
	return nil
}

// NewMeter creates a meter and registers it with the receiver.
func (r *Registry) NewMeter(tyep interface{}, name string) *Meter {
	m := newMeter()
	if r.register(tyep, name, m) {
		return m
	}
	return nil
}

// NewDistribuion creates a distribution and registers it with the receiver.
func (r *Registry) NewDistribution(tyep interface{},
	name string) *Distribution {

	m := newDistribution()
	if r.register(tyep, name, m) {
		return m
	}
	return nil
}

// NewGauge creates a gauge and registers it with the receiver.
func (r *Registry) NewGauge(tyep interface{}, name string) *Gauge {
	m := newGauge()
	if r.register(tyep, name, m) {
		return m
	}
	return nil
}

func (r *Registry) Name() string {
	r.lock.RLock() // probably unnecessary
	ret := r.name
	r.lock.RUnlock()
	return ret
}

// List returns a slice of arrays. Each array contains a metric name at
// the zeroth index and its type in the first.
func (r *Registry) List() [][2]string {
	r.lock.RLock()
	list := make([][2]string, len(r.metrics))
	i := 0
	for name, metric := range r.metrics {
		list[i][0] = name
		var t string
		switch metric.(type) {
		case *Counter:
			t = "counter"
		case *Distribution:
			t = "distribution"
		case *Gauge:
			t = "gauge"
		case *Meter:
			t = "meter"
		}
		list[i][1] = t
		i++
	}
	r.lock.RUnlock()
	return list
}

// ListMetrics returns a map with keys representing metric names
// and values containing the corresponding metric.
func (r *Registry) ListMetrics() map[string]Metric {
	list := make(map[string]Metric)
	r.lock.RLock()
	for name, metric := range r.metrics {
		list[name] = metric
	}
	r.lock.RUnlock()
	return list
}

func (r *Registry) Find(tyep interface{}, name string) Metric {
	typeName := realType(tyep)
	return r.FindS(fmt.Sprintf("%s.%s", typeName, name))
}

func (r *Registry) FindS(fullname string) Metric {
	r.lock.RLock()
	ret := r.metrics[fullname]
	r.lock.RUnlock()
	return ret
}

var DefaultRegistry *Registry

func init() {
	DefaultRegistry = NewRegistry("default")
}

// NewCounter creates a counter and registers it with the 
// default registry.
func NewCounter(tyep interface{}, name string) *Counter {
	return DefaultRegistry.NewCounter(tyep, name)
}

// NewDistribution creates a distribution and registers it with the 
// default registry.
func NewDistribution(tyep interface{}, name string) *Distribution {
	return DefaultRegistry.NewDistribution(tyep, name)
}

// NewGauge creates a gauge and registers it with the 
// default registry.
func NewGauge(tyep interface{}, name string) *Gauge {
	return DefaultRegistry.NewGauge(tyep, name)
}

// NewMeter creates a meter and registers it with the 
// default registry.
func NewMeter(tyep interface{}, name string) *Meter {
	return DefaultRegistry.NewMeter(tyep, name)
}
