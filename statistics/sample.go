package statistics

import (
	"math"
	"metrics/rbtree"
)

type Sample struct {
	values *rbtree.Tree

	mean           float64
	secondCMtimesN float64
	thirdCMtimesN  float64
	fourthCMtimesN float64
}

type SampleElement struct {
	node *rbtree.Node
}

func NewSample() *Sample {
	return &Sample{
		values: rbtree.New(),
	}
}

func (s *Sample) Count() uint64 {
	return s.values.Size()
}

func (s *Sample) Add(v int64) SampleElement {
	node := s.values.Insert(v, nil)

	x := float64(v)
	n := float64(s.Count())

	// http://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
	delta := x - s.mean
	deltaOverN := delta / n
	a := delta * deltaOverN * deltaOverN * (n - 1.0)
	b := deltaOverN * s.secondCMtimesN

	s.mean += deltaOverN

	s.fourthCMtimesN += deltaOverN * a * (n*n - 3.0*n + 3.0)
	s.fourthCMtimesN += 6.0 * deltaOverN * b
	s.fourthCMtimesN -= 4.0 * deltaOverN * s.thirdCMtimesN

	s.thirdCMtimesN += a * (n - 2.0)
	s.thirdCMtimesN -= 3.0 * b

	s.secondCMtimesN += delta * deltaOverN * (n - 1.0)

	return SampleElement{node}
}

func (s *Sample) Remove(se SampleElement) {
	defer s.values.RemoveNode(se.node)

	if s.Count() <= 1 {
		s.mean = 0
		s.secondCMtimesN = 0
		s.thirdCMtimesN = 0
		s.fourthCMtimesN = 0
		return
	}

	n := float64(s.Count()) // n is at least 2
	x := float64(se.node.Key())

	delta := n / (n - 1.0) * (x - s.mean)
	deltaOverN := delta / n
	a := delta * deltaOverN * deltaOverN * (n - 1.0)

	s.mean -= deltaOverN

	s.secondCMtimesN -= delta * deltaOverN * (n - 1.0)

	b := deltaOverN * s.secondCMtimesN

	s.thirdCMtimesN -= delta * deltaOverN * deltaOverN * (n - 1.0) * (n - 2.0)
	s.thirdCMtimesN += 3.0 * deltaOverN * s.secondCMtimesN

	s.fourthCMtimesN -= deltaOverN * a * (n*n - 3.0*n + 3.0)
	s.fourthCMtimesN -= 6.0 * deltaOverN * b
	s.fourthCMtimesN += 4.0 * deltaOverN * s.thirdCMtimesN
}

func (s *Sample) Mean() float64 {
	return toFinite(s.mean)
}

func (s *Sample) Variance() float64 {
	return toFinite(s.secondCMtimesN / float64(s.Count()-1))
}

func (s *Sample) StandardDeviation() float64 {
	return toFinite(math.Sqrt(s.Variance()))
}

func (s *Sample) Skewness() float64 {
	r := math.Sqrt(float64(s.Count())) * s.thirdCMtimesN
	v := r / math.Pow(s.secondCMtimesN, 1.5)
	return toFinite(v)
}

func (s *Sample) Kurtosis() float64 {
	r := float64(s.Count()) * s.fourthCMtimesN
	v := r / s.secondCMtimesN / s.secondCMtimesN
	return toFinite(v)
}

func (s *Sample) Percentile(p float64) int64 {
	if p < 0.0 {
		p = 0.0
	}
	if p > 1.0 {
		p = 1.0
	}
	//math.Floor is needed to work around 6g issue 3804
	//f64 -> u64 does not truncate
	rank := uint64(math.Floor(p*float64(s.Count()-1) + 0.5))

	node := s.values.FindByRank(rank)
	if node == nil {
		return 0
	}
	return node.Key()
}

func (se SampleElement) Value() int64 {
	return se.node.Key()
}

func toFinite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
