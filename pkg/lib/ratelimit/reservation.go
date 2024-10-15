package ratelimit

import (
	"time"
)

type Reservation struct {
	key        string
	spec       BucketSpec
	ok         bool
	err        error
	tokenTaken int
	timeToAct  *time.Time
	isConsumed bool
}

func (r *Reservation) Error() error {
	if r == nil {
		return nil
	}
	if r.err != nil {
		return r.err
	}
	if !r.ok {
		return ErrRateLimited(r.spec.Name)
	}
	return nil
}

func (r *Reservation) GetTimeToAct() time.Time {
	if r == nil || r.timeToAct == nil {
		return time.Unix(0, 0).UTC()
	}
	return *r.timeToAct
}

func (r *Reservation) Consume() {
	if r == nil {
		return
	}
	r.isConsumed = true
}
