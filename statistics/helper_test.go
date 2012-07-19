package statistics

import (
	"math"
	"testing"
	"time"
)

var testSampleSet = [50]int64{
	9675058, -1689853, -3725820, -705873, -3251762, -4231217, 1198474,
	-1222771, 6042688, 1417426, 9394784, -3796327, 1215914, 4205163,
	-8477723, 3520070, 745446, 8757462, -7131680, -6519221, -8401375,
	-1795469, -5918478, -6614860, -2001987, -9988298, -626305, -7724919,
	9694132, 5006064, 7279687, -1673061, -9803177, 6115289, 647511,
	3251507, -4252489, -1598969, 4168172, 9554726, -4176556, 8863435,
	1681047, -169245, 912269, -9210523, 4355342, 8089016, 7113846, 7384336,
}

var testTimeOffsets = [50]time.Duration{
	3415, 4722, 5704, 9097, 15862, 16712, 16967, 18683, 19004, 19653,
	20430, 20438, 22254, 22556, 23400, 24402, 24752, 26056, 29225, 31666,
	32095, 39859, 46793, 48826, 48900, 57867, 58500, 59443, 59580, 61345,
	63637, 65565, 68570, 68612, 69711, 72196, 72247, 73067, 75216, 75252,
	76043, 79518, 79686, 80669, 85560, 87537, 88372, 93209, 97658, 98853,
}

const (
	testAlmostEqualTolerance = 1e-13
)

func testCompare(t *testing.T, name string, a interface{}, b interface{}) {
	var eq bool
	switch a.(type) {
	case float64:
		eq = testAlmostEqual(a.(float64), b.(float64))
	default:
		eq = (a == b)
	}
	if !eq {
		t.Errorf("%s is %v, should be %v.", name, a, b)
	}
}

func testAlmostEqual(a float64, b float64) bool {
	diff := math.Abs(math.Abs(a/b - 1))
	return diff < testAlmostEqualTolerance
}
