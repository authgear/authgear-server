package lockout

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const untilKey = "until"

var ErrLocked = apierrors.TooManyRequest.WithReason("AccountLockout")

func NewErrLocked(until time.Time) error {
	return ErrLocked.NewWithInfo("account locked", apierrors.Details{
		untilKey: until,
	})
}
