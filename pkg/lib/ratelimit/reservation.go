package ratelimit

import (
	"time"
)

type Reservation struct {
	Key                string     `json:"key"`
	Spec               BucketSpec `json:"spec"`
	TokenTaken         float64    `json:"token_taken"`
	WasCancelPrevented bool       `json:"was_cancel_prevented"`
	IsCancelled        bool       `json:"is_cancelled"`
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
	r.WasCancelPrevented = true
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
	return ErrRateLimited(r.spec.Name)
}

func (r *FailedReservation) GetTimeToAct() time.Time {
	if r == nil {
		return time.Unix(0, 0).UTC()
	}
	return r.timeToAct
}
