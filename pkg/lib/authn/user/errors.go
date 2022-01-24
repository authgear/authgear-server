package user

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserNotFound = errors.New("user not found")

var DisabledUser = apierrors.Forbidden.WithReason("DisabledUser")

func NewErrDisabledUser(reason *string) error {
	return DisabledUser.NewWithInfo("user is disabled", map[string]interface{}{
		"reason": reason,
	})
}

var ErrDeactivatedUser = apierrors.Forbidden.WithReason("DeactivatedUser").New("user is deactivated")

var ErrScheduledDeletionByAdmin = apierrors.Forbidden.WithReason("ScheduleDeletionByAdmin").New("user was scheduled for deletion by admin")
var ErrScheduledDeletionByEndUser = apierrors.Forbidden.WithReason("ScheduledDeletionByEndUser").New("user was scheduled for deletion by end-user")
