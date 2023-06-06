package lockout

import (
	"time"
)

type Storage interface {
	Update(spec BucketSpec, delta int) (lockedUntil *time.Time, err error)
	Clear(spec BucketSpec) error
}
