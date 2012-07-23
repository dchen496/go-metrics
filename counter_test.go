package metrics

import (
	"testing"
)

func testCounterInit() *Counter {
	c := newCounter()
	c.Set(1357)
	return c
}

func TestCounterInc(t *testing.T) {
	c := testCounterInit()
	c.Inc(10)
	s := c.Snapshot()
	if s.Value != 1367 {
		t.Errorf("Counter incremented to %d, expected %d", s.Value, 1367)
	}
}
