package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var errTooManyRequestsMsg = "request rate limited"
var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

var ErrTooManyRequests = RateLimited.New(errTooManyRequestsMsg)
var ErrUsageLimitExceeded = apierrors.ServiceUnavailable.WithReason("UsageLimitExceeded").
	New("usage limit exceeded")

func ErrTooManyRequestsFrom(bucket Bucket) error {
	if bucket.Name == "" {
		return ErrTooManyRequests
	} else {
		return RateLimited.NewWithInfo(errTooManyRequestsMsg, apierrors.Details{
			"bucket_name": bucket.Name,
		})
	}
}
