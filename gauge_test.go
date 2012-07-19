package metrics

import (
	"fmt"
	"testing"
)

type testGaugable struct {
	value  uint64
	status bool
}

func (t *testGaugable) String() string {
	return fmt.Sprint(t.value, t.status)
}

func testGaugeInit() *Gauge {
	g := newGauge()
	t := testGaugable{}
	g.SetFunction(func(*Gauge) Gaugable {
		t.value += 5
		t.status = true
		return &t
	})
	g.update(testTime)
	return g
}

func TestGaugeUpdate(t *testing.T) {
	g := testGaugeInit()

	expected := fmt.Sprint(5, true)
	s := g.Snapshot()
	value := s.Value.String()
	if value != expected {
		t.Errorf("Wrong data in Gauge.Value after autoupdate: expected %s, got %s",
			expected, value)
	}
	if s.LastUpdated != testTime {
		t.Errorf("Counter updated time is to %v, expected %v",
			s.LastUpdated, testTime)
	}
}