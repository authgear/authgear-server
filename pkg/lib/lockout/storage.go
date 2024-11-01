package lockout

import (
	"context"
	"time"
)

type Storage interface {
	Update(ctx context.Context, spec LockoutSpec, contributor string, delta int) (isSuccess bool, lockedUntil *time.Time, err error)
	Clear(ctx context.Context, spec LockoutSpec, contributor string) error
}
