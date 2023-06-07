package lockout

import (
	"time"
)

type Storage interface {
	Update(spec BucketSpec, delta int) (isSuccess bool, lockedUntil *time.Time, err error)
	Clear(spec BucketSpec) error
}
