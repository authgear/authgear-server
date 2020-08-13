package clock

import "time"

type Clock interface {
	NowUTC() time.Time
	NowMonotonic() time.Time
}

type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (*SystemClock) NowUTC() time.Time {
	return time.Now().UTC()
}

func (*SystemClock) NowMonotonic() time.Time {
	return time.Now()
}
