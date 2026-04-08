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
