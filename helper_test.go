package metrics

import (
	"time"
)

var testTime time.Time

func init() {
	testTime = time.Date(1970, 1, 1, 1, 1, 1, 1, time.UTC)
}
