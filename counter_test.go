package metrics

import (
	"testing"
	"time"
)

func testCounterInit() *Counter {
	c := newCounter()
	c.set(1357, testTime)
	return c
}

func TestCounterInc(t *testing.T) {
	c := testCounterInit()
	newTime := testTime.Add(time.Duration(100))
	c.inc(10, newTime)
	s := c.Snapshot()
	if s.Value != 1367 {
		t.Errorf("Counter incremented to %d, expected %d", s.Value, 1367)
	}
	if s.LastUpdated != newTime {
		t.Errorf("Counter updated time is to %v, expected %v",
			s.LastUpdated, newTime)
	}
}

func BenchmarkCounterUpdate(b *testing.B) {
	c := newCounter()
	for i := 0; i < b.N; i++ {
		c.inc(int64(i), testTime.Add(time.Duration(i)))
	}
}
