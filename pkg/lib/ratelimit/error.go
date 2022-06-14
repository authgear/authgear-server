package ratelimit

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrTooManyRequests = apierrors.TooManyRequest.WithReason("RateLimited").
	New("request rate limited")

var ErrUsageLimitExceeded = apierrors.ServiceUnavailable.WithReason("UsageLimitExceeded").
	New("usage limit exceeded")
