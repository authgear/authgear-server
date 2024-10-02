package usage

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var UsageLimitExceeded = apierrors.TooManyRequest.WithReason("UsageLimitExceeded")

func ErrUsageLimitExceeded(name LimitName) error {
	return UsageLimitExceeded.NewWithInfo("usage limit exceeded", apierrors.Details{
		"name": name,
	})
}
