package statistics

import (
	"time"
)

type Rate struct {
	value         int64
	lastUpdated   time.Time
	timeConstants []time.Duration
	derivatives   [][]float64
}

func NewRate(derivatives uint64, tcs []time.Duration) *Rate {
	r := Rate{}
	r.allocate(derivatives, uint64(len(tcs)))
	copy(r.timeConstants, tcs)
	return &r
}

func (r *Rate) SetTimeConstants(tcs []time.Duration) {
	r.allocate(uint64(len(r.derivatives)-1), uint64(len(tcs)))
	copy(r.timeConstants, tcs)
}

func (r *Rate) SetMaxDerivativeOrder(n uint64) {
	r.allocate(n, uint64(len(r.timeConstants)))
}

func (r *Rate) allocate(derivatives uint64, times uint64) {
	r.timeConstants = make([]time.Duration, times)
	r.derivatives = make([][]float64, derivatives+1)
	for order := range r.derivatives {
		r.derivatives[order] = make([]float64, times+1)
	}
}

func (r *Rate) Reset() {
	r.value = 0
	r.lastUpdated = time.Time{}
	for i := 0; i < len(r.derivatives); i++ {
		for j := 0; j < len(r.derivatives[i]); j++ {
			r.derivatives[i][j] = 0
		}
	}
}

func (r *Rate) Value() int64 {
	return r.value
}

func (r *Rate) LastUpdated() time.Time {
	return r.lastUpdated
}

func (r *Rate) Set(v int64, t time.Time) {
	if !t.After(r.lastUpdated) {
		return
	}

	old := r.derivatives[0][0]
	r.derivatives[0][0] = float64(v)

	if !r.lastUpdated.IsZero() {
		dt := float64(t.Sub(r.lastUpdated)) / float64(time.Second)
		for i := 1; i < len(r.derivatives); i++ {
			old, r.derivatives[i][0] = r.derivatives[i][0],
				(r.derivatives[i-1][0]-old)/dt
		}

		for tcInd, tc := range r.timeConstants {
			k := float64(tc) / float64(tc+t.Sub(r.lastUpdated))
			for order := range r.derivatives {
				r.derivatives[order][tcInd+1] *= k
				r.derivatives[order][tcInd+1] += (1.0 - k) * r.derivatives[order][0]
			}
		}
	}

	r.value = v
	r.lastUpdated = t
}

func (r *Rate) TimeConstants() []time.Duration {
	m := make([]time.Duration, len(r.timeConstants))
	copy(m, r.timeConstants)
	return m
}

// zeroth time constant is the instantaneous rate of change,
// the rest are indexed starting from 1
func (r *Rate) Derivatives() [][]float64 {
	m := make([][]float64, len(r.derivatives))
	for i := range m {
		m[i] = make([]float64, len(r.derivatives[i]))
		copy(m[i], r.derivatives[i])
	}
	return m
}
