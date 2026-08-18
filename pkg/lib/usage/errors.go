package usage

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

var UsageLimitExceeded = apierrors.TooManyRequest.WithReason("UsageLimitExceeded")

func ErrUsageLimitExceeded(name model.UsageName, period model.UsageLimitPeriod) error {
	return UsageLimitExceeded.NewWithInfo("usage limit exceeded", apierrors.Details{
		// name is kept for backward compatibility.
		"name":       legacyLimitName(name),
		"usage_name": name,
		"period":     period,
	})
}

// ErrStandingUsageLimitExceeded is the standing-limit counterpart to
// ErrUsageLimitExceeded. It does not go through legacyLimitName, which
// panics on any model.UsageName it doesn't have a case for — a standing
// limit has no Redis key/legacy name at all.
func ErrStandingUsageLimitExceeded(name model.UsageName) error {
	return UsageLimitExceeded.NewWithInfo("usage limit exceeded", apierrors.Details{
		"usage_name": name,
	})
}
