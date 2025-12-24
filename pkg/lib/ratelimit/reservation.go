package ratelimit

import (
	"time"
)

type Reservation struct {
	key                string
	spec               BucketSpec
	tokenTaken         float64
	wasCancelPrevented bool
}

// PreventCancel prevents r from being Cancel().
// The typical usage is like
// r := ...
// defer Cancel(r)
// ...
// Discover a situation that r must not be canceled.
// r.PreventCancel()
func (r *Reservation) PreventCancel() {
	if r == nil {
		return
	}
	r.wasCancelPrevented = true
}

type FailedReservation struct {
	key       string
	spec      BucketSpec
	timeToAct time.Time
}

func (r *FailedReservation) Error() error {
	if r == nil {
		return nil
	}
	return ErrRateLimited(r.spec.RateLimitName, r.spec.RateLimitGroup, r.spec.Name)
}

func (r *FailedReservation) GetTimeToAct() time.Time {
	if r == nil {
		return time.Unix(0, 0).UTC()
	}
	return r.timeToAct
}

// For test case only
func NewFailedReservation(spec BucketSpec) *FailedReservation {
	return &FailedReservation{
		spec: spec,
	}
}
