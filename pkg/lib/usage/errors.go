package usage

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
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
//
// quota is carried so a caller building an audit-event payload (CIMD's
// oauth.client.resolution.failed, DCR's oauth.client.registration.failed)
// can report which limit was hit without a second lookup -- see
// StandingUsageLimitDetails.
func ErrStandingUsageLimitExceeded(name model.UsageName, quota int) error {
	return UsageLimitExceeded.NewWithInfo("usage limit exceeded", apierrors.Details{
		"usage_name": name,
		"quota":      quota,
	})
}

// StandingUsageLimitDetails extracts the usage name and quota from an error
// returned by Limiter.CheckStanding, for an audit-event payload that needs
// to report which limit was hit. ok is false if err is not one of these
// errors (e.g. a plain infrastructure error), in which case name/quota are
// the zero value and must not be used.
func StandingUsageLimitDetails(err error) (name model.UsageName, quota int, ok bool) {
	var apiErr *apierrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Kind != UsageLimitExceeded {
		return "", 0, false
	}
	// NewWithInfo wraps every value in errorutil.DetailTaggedValue (so it
	// renders in the JSON response); FilterDetails is the established way
	// to unwrap those back to their plain values.
	details := errorutil.FilterDetails(apiErr.Info_ReadOnly, apierrors.APIErrorDetail)
	name, _ = details["usage_name"].(model.UsageName)
	quota, _ = details["quota"].(int)
	return name, quota, true
}
