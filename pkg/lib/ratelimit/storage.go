package ratelimit

import (
	"context"
	"time"
)

type Storage interface {
	Update(ctx context.Context, key string, period time.Duration, burst int, delta int) (ok bool, timeToAct time.Time, err error)
}
