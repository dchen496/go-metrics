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
	if s.Value() != 1367 {
		t.Errorf("Counter incremented to %d, expected %d", s.Value(), 1367)
	}
	if s.LastUpdated() != newTime {
		t.Errorf("Counter updated time is to %v, expected %v",
			s.LastUpdated(), newTime)
	}
}

func TestCounterProcess(t *testing.T) {
	c := testCounterInit()
	opt := &CounterProcessOptions{Derivatives: true}
	tp := &testProcessor{}
	out := c.Process(tp, "test", opt)
	switch out.(int) {
	case 1:
	case -1:
		t.Errorf("Counter processor failed, wrong name")
	case -2:
		t.Errorf("Counter processor failed, wrong value")
	case -3:
		t.Errorf("Counter processor failed, wrong set of options")
	case -4:
		t.Errorf("Counter processor failed, wrong time")
	case 2, 3:
		t.Errorf("Counter processor failed, wrong processor")
	default:
		t.Errorf("Counter processor failed, expected %v, got %v", 1, out)
	}
}

func BenchmarkCounterUpdate(b *testing.B) {
	c := newCounter()
	for i := 0; i < b.N; i++ {
		c.inc(int64(i), testTime.Add(time.Duration(i)))
	}
}
