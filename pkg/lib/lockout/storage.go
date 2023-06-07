package lockout

import (
	"time"
)

type Storage interface {
	Update(spec BucketSpec, contributor string, delta int) (isSuccess bool, lockedUntil *time.Time, err error)
	Clear(spec BucketSpec, contributor string) error
}
