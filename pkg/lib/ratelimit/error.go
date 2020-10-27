package ratelimit

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrTooManyRequests = apierrors.TooManyRequest.WithReason("RateLimited").
	New("request rate limited")
