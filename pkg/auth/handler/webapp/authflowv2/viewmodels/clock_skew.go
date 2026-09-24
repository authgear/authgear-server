package viewmodels

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

type ClockSkewViewModel struct {
	ClockSkewCheckEnabled bool
	ServerTimeUnixMilli   int64
	// ClockSkewMaxAheadMilli and ClockSkewMaxBehindMilli mirror the DPoP proof iat validation in pkg/lib/dpop.
	ClockSkewMaxAheadMilli  int64
	ClockSkewMaxBehindMilli int64
}

// A token request with a DPoP proof issued by a skewed device clock will fail,
// so the check only matters when the authorization request is DPoP-bound.
func NewClockSkewViewModel(dpopEnabled bool, now time.Time) ClockSkewViewModel {
	return ClockSkewViewModel{
		ClockSkewCheckEnabled:   dpopEnabled,
		ServerTimeUnixMilli:     now.UnixMilli(),
		ClockSkewMaxAheadMilli:  duration.ClockSkew.Milliseconds(),
		ClockSkewMaxBehindMilli: duration.Short.Milliseconds(),
	}
}
