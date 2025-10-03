package user

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type AccountStatusType string

const (
	AccountStatusTypeNormal                         AccountStatusType = "normal"
	AccountStatusTypeDisabled                       AccountStatusType = "disabled"
	AccountStatusTypeDeactivated                    AccountStatusType = "deactivated"
	AccountStatusTypeScheduledDeletionDisabled      AccountStatusType = "scheduled_deletion_disabled"
	AccountStatusTypeScheduledDeletionDeactivated   AccountStatusType = "scheduled_deletion_deactivated"
	AccountStatusTypeAnonymized                     AccountStatusType = "anonymized"
	AccountStatusTypeScheduledAnonymizationDisabled AccountStatusType = "scheduled_anonymization_disabled"
)

// AccountStatus represents disabled, deactivated, or scheduled deletion state.
// The zero value means normal.
type AccountStatus struct {
	IsDisabled               bool
	AccountStatusStaleFrom   *time.Time
	IsIndefinitelyDisabled   *bool
	IsDeactivated            *bool
	DisableReason            *string
	TemporarilyDisabledFrom  *time.Time
	TemporarilyDisabledUntil *time.Time
	AccountValidFrom         *time.Time
	AccountValidUntil        *time.Time
	DeleteAt                 *time.Time
	AnonymizeAt              *time.Time
	AnonymizedAt             *time.Time
	IsAnonymized             *bool
}

func (s AccountStatus) Type() AccountStatusType {
	if !s.IsDisabled {
		return AccountStatusTypeNormal
	}
	if s.DeleteAt != nil {
		if s.IsDeactivated != nil && *s.IsDeactivated {
			return AccountStatusTypeScheduledDeletionDeactivated
		}
		return AccountStatusTypeScheduledDeletionDisabled
	}
	if s.IsAnonymized != nil && *s.IsAnonymized {
		return AccountStatusTypeAnonymized
	}
	if s.AnonymizeAt != nil {
		return AccountStatusTypeScheduledAnonymizationDisabled
	}
	if s.IsDeactivated != nil && *s.IsDeactivated {
		return AccountStatusTypeDeactivated
	}
	return AccountStatusTypeDisabled
}

func (s AccountStatus) Check() error {
	// This method must be in sync with IsAccountStatusError.
	type_ := s.Type()
	switch type_ {
	case AccountStatusTypeNormal:
		return nil
	case AccountStatusTypeDisabled:
		return NewErrDisabledUser(s.DisableReason)
	case AccountStatusTypeDeactivated:
		return ErrDeactivatedUser
	case AccountStatusTypeAnonymized:
		return ErrAnonymizedUser
	case AccountStatusTypeScheduledDeletionDisabled:
		return NewErrScheduledDeletionByAdmin(*s.DeleteAt)
	case AccountStatusTypeScheduledDeletionDeactivated:
		return NewErrScheduledDeletionByEndUser(*s.DeleteAt)
	case AccountStatusTypeScheduledAnonymizationDisabled:
		return NewErrScheduledAnonymizationByAdmin(*s.AnonymizeAt)
	default:
		panic(fmt.Errorf("unknown account status type: %v", type_))
	}
}

func (s AccountStatus) Reenable() (*AccountStatus, error) {
	false_ := false

	target := s
	target.IsDisabled = false
	target.IsIndefinitelyDisabled = &false_
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = nil
	target.IsAnonymized = &false_

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.Type() == AccountStatusTypeDisabled {
		return &target, nil
	}

	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) Disable(reason *string) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.IsDisabled = true
	target.IsIndefinitelyDisabled = &true_
	target.IsDeactivated = &false_
	target.DisableReason = reason
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = nil
	target.IsAnonymized = &false_

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleDeletionByEndUser(deleteAt time.Time) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.IsDisabled = true
	target.IsIndefinitelyDisabled = &true_
	target.IsDeactivated = &true_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = &deleteAt
	target.AnonymizeAt = nil
	target.IsAnonymized = &false_

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.Type() != AccountStatusTypeNormal {
		return nil, s.makeTransitionError(target.Type())
	}

	return &target, nil
}

func (s AccountStatus) ScheduleDeletionByAdmin(deleteAt time.Time) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.IsDisabled = true
	target.IsIndefinitelyDisabled = &true_
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = &deleteAt
	target.AnonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.DeleteAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleDeletionByAdmin() (*AccountStatus, error) {
	isAnonymized := false
	if s.IsAnonymized != nil && *s.IsAnonymized {
		isAnonymized = true
	}
	false_ := false

	target := s
	target.IsDisabled = isAnonymized
	target.IsIndefinitelyDisabled = &isAnonymized
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.DeleteAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) Anonymize(now time.Time) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.IsDisabled = true
	target.IsIndefinitelyDisabled = &true_
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = nil
	target.IsAnonymized = &true_
	target.AnonymizedAt = &now

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleAnonymizationByAdmin(anonymizeAt time.Time) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.IsDisabled = true
	target.IsIndefinitelyDisabled = &true_
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = &anonymizeAt
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.AnonymizeAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleAnonymizationByAdmin() (*AccountStatus, error) {
	false_ := false

	target := s
	target.IsDisabled = false
	target.IsIndefinitelyDisabled = &false_
	target.IsDeactivated = &false_
	target.DisableReason = nil
	target.TemporarilyDisabledFrom = nil
	target.TemporarilyDisabledUntil = nil
	target.DeleteAt = nil
	target.AnonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.AnonymizeAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) makeTransitionError(targetType AccountStatusType) error {
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", s.Type(), targetType),
		map[string]interface{}{
			"from": s.Type(),
			"to":   targetType,
		},
	)
}

func IsAccountStatusError(err error) bool {
	// This function must be in sync with AccountStatus.Check.
	switch {
	case apierrors.IsKind(err, DisabledUser):
		return true
	case apierrors.IsKind(err, DeactivatedUser):
		return true
	case apierrors.IsKind(err, AnonymizedUser):
		return true
	case apierrors.IsKind(err, ScheduledDeletionByAdmin):
		return true
	case apierrors.IsKind(err, ScheduledDeletionByEndUser):
		return true
	case apierrors.IsKind(err, ScheduledAnonymizationByAdmin):
		return true
	default:
		return false
	}
}
