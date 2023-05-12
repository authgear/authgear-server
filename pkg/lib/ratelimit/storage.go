package ratelimit

import (
	"time"
)

type Storage interface {
	Update(spec BucketSpec, delta int) (ok bool, timeToAct time.Time, err error)
}
