package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUsageLimitExceeded = apierrors.ServiceUnavailable.WithReason("UsageLimitExceeded").
	New("usage limit exceeded")

func ErrTooManyRequestsFrom(bucket Bucket) error {
	RateLimited := apierrors.TooManyRequest.WithReason("RateLimited")
	errMsg := "request rate limited"
	if bucket.Name == "" {
		return RateLimited.New(errMsg)
	} else {
		return RateLimited.NewWithInfo(errMsg, apierrors.Details{
			"bucket_name": bucket.Name,
		})
	}
}
