package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUsageLimitExceeded = apierrors.ServiceUnavailable.WithReason("UsageLimitExceeded").
	New("usage limit exceeded")

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrTooManyRequestsFrom(bucket Bucket) error {
	errMsg := "request rate limited"
	if bucket.Name == "" {
		return RateLimited.New(errMsg)
	} else {
		return RateLimited.NewWithInfo(errMsg, apierrors.Details{
			"bucket_name": bucket.Name,
		})
	}
}
