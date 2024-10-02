package ratelimit

import (
	"time"
)

type Storage interface {
	Update(key string, period time.Duration, burst int, delta int) (ok bool, timeToAct time.Time, err error)
}
