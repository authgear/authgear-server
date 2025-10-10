package lockout

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const untilKey = "until"

var AccountLockout = apierrors.TooManyRequest.WithReason("AccountLockout")

func NewErrLocked(until time.Time) error {
	return AccountLockout.NewWithInfo("account locked", apierrors.Details{
		untilKey: until,
	})
}
