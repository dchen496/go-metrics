package metrics

import (
	"testing"
	"time"
)

func testMeterInit() *Meter {
	m := newMeter()
	m.set(1357, testTime)
	return m
}

func TestMeterInc(t *testing.T) {
	m := testMeterInit()
	newTime := testTime.Add(time.Duration(100))
	m.inc(10, newTime)
	s := m.Snapshot()
	if s.Value != 1367 {
		t.Errorf("Meter incremented to %d, expected %d", s.Value, 1367)
	}
	if s.LastUpdated != newTime {
		t.Errorf("Meter updated time is to %v, expected %v",
			s.LastUpdated, newTime)
	}
}

func BenchmarkMeterUpdate(b *testing.B) {
	m := newMeter()
	for i := 0; i < b.N; i++ {
		m.inc(int64(i), testTime.Add(time.Duration(i)))
	}
}
