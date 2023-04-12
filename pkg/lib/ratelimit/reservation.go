package ratelimit

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type Reservation struct {
	spec       BucketSpec
	ok         bool
	err        error
	tokenTaken int
	timeToAct  *time.Time
}

func (r *Reservation) Error() error {
	if r.err != nil {
		return r.err
	}
	if !r.ok {
		return RateLimited.NewWithInfo("request rate limited", apierrors.Details{
			bucketNameKey: r.spec.Name,
		})
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
