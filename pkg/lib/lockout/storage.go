package lockout

import (
	"time"
)

type Storage interface {
	Update(spec LockoutSpec, contributor string, delta int) (isSuccess bool, lockedUntil *time.Time, err error)
	Clear(spec LockoutSpec, contributor string) error
}
