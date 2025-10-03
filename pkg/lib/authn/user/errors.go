package user

import (
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserNotFound = errors.New("user not found")

var DisabledUser = apierrors.Forbidden.WithReason("DisabledUser")
var DeactivatedUser = apierrors.Forbidden.WithReason("DeactivatedUser")
var AnonymizedUser = apierrors.Forbidden.WithReason("AnonymizedUser")
var ScheduledDeletionByAdmin = apierrors.Forbidden.WithReason("ScheduledDeletionByAdmin")
var ScheduledDeletionByEndUser = apierrors.Forbidden.WithReason("ScheduledDeletionByEndUser")
var ScheduledAnonymizationByAdmin = apierrors.Forbidden.WithReason("ScheduledAnonymizationByAdmin")

var InvalidAccountStatusTransition = apierrors.Invalid.WithReason("InvalidAccountStatusTransition")

func NewErrDisabledUser(reason *string) error {
	return DisabledUser.NewWithInfo("user is disabled", map[string]interface{}{
		"reason": reason,
	})
}

var ErrDeactivatedUser = DeactivatedUser.New("user is deactivated")
var ErrAnonymizedUser = AnonymizedUser.New("user is anonymized")

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

func NewErrScheduledAnonymizationByAdmin(anonymizeAt time.Time) error {
	return ScheduledAnonymizationByAdmin.NewWithInfo("user was scheduled for anonymization by admin", map[string]interface{}{
		"anonymize_at": anonymizeAt,
	})
}
