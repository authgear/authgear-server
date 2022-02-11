package user

import (
	"errors"
	"time"

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

var ScheduledDeletionByAdmin = apierrors.Forbidden.WithReason("ScheduledDeletionByAdmin")
var ScheduledDeletionByEndUser = apierrors.Forbidden.WithReason("ScheduledDeletionByEndUser")

func NewErrScheduledDeletionByAdmin(deleteAt time.Time) error {
	return ScheduledDeletionByAdmin.NewWithInfo("user was scheduled for deletion by admin", map[string]interface{}{
		"delete_at": deleteAt,
	})
}

func NewErrScheduledDeletionByEndUser(deleteAt time.Time) error {
	return ScheduledDeletionByEndUser.NewWithInfo("user was scheduled for deletion by end-user", map[string]interface{}{
		"delete_at": deleteAt,
	})
}
