package ratelimit

import (
	"time"
)

type Reservation struct {
	spec       BucketSpec
	ok         bool
	err        error
	tokenTaken int
	timeToAct  *time.Time
	isConsumed bool
}

func (r *Reservation) Error() error {
	if r.err != nil {
		return r.err
	}
	if !r.ok {
		return ErrRateLimited(r.spec.Name)
	}
	return nil
}

func (r *Reservation) DelayFrom(t time.Time) time.Duration {
	if r.timeToAct == nil {
		return 0
	}
	delay := r.timeToAct.Sub(t)
	if delay < 0 {
		return 0
	}
	return delay
}

func (r *Reservation) Consume() {
	r.isConsumed = true
}
