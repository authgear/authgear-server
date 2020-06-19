package clock

import "time"

type Clock interface {
	NowUTC() time.Time
	Now() time.Time
}

type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (*SystemClock) NowUTC() time.Time {
	return time.Now().UTC()
}

func (*SystemClock) Now() time.Time {
	return time.Now()
}
