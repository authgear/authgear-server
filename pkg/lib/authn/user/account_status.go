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
	isDisabled               bool
	accountStatusStaleFrom   *time.Time
	isIndefinitelyDisabled   *bool
	isDeactivated            *bool
	disableReason            *string
	temporarilyDisabledFrom  *time.Time
	temporarilyDisabledUntil *time.Time
	accountValidFrom         *time.Time
	accountValidUntil        *time.Time
	deleteAt                 *time.Time
	anonymizeAt              *time.Time
	anonymizedAt             *time.Time
	isAnonymized             *bool
}

func (s AccountStatus) Type() AccountStatusType {
	if !s.isDisabled {
		return AccountStatusTypeNormal
	}
	if s.deleteAt != nil {
		if s.isDeactivated != nil && *s.isDeactivated {
			return AccountStatusTypeScheduledDeletionDeactivated
		}
		return AccountStatusTypeScheduledDeletionDisabled
	}
	if s.isAnonymized != nil && *s.isAnonymized {
		return AccountStatusTypeAnonymized
	}
	if s.anonymizeAt != nil {
		return AccountStatusTypeScheduledAnonymizationDisabled
	}
	if s.isDeactivated != nil && *s.isDeactivated {
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
		return NewErrDisabledUser(s.disableReason)
	case AccountStatusTypeDeactivated:
		return ErrDeactivatedUser
	case AccountStatusTypeAnonymized:
		return ErrAnonymizedUser
	case AccountStatusTypeScheduledDeletionDisabled:
		return NewErrScheduledDeletionByAdmin(*s.deleteAt)
	case AccountStatusTypeScheduledDeletionDeactivated:
		return NewErrScheduledDeletionByEndUser(*s.deleteAt)
	case AccountStatusTypeScheduledAnonymizationDisabled:
		return NewErrScheduledAnonymizationByAdmin(*s.anonymizeAt)
	default:
		panic(fmt.Errorf("unknown account status type: %v", type_))
	}
}

func (s AccountStatus) Reenable() (*AccountStatus, error) {
	false_ := false

	target := s
	target.isDisabled = false
	target.isIndefinitelyDisabled = &false_
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = nil
	target.isAnonymized = &false_

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
	target.isDisabled = true
	target.isIndefinitelyDisabled = &true_
	target.isDeactivated = &false_
	target.disableReason = reason
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = nil
	target.isAnonymized = &false_

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
	target.isDisabled = true
	target.isIndefinitelyDisabled = &true_
	target.isDeactivated = &true_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = &deleteAt
	target.anonymizeAt = nil
	target.isAnonymized = &false_

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
	target.isDisabled = true
	target.isIndefinitelyDisabled = &true_
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = &deleteAt
	target.anonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.deleteAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleDeletionByAdmin() (*AccountStatus, error) {
	isAnonymized := false
	if s.isAnonymized != nil && *s.isAnonymized {
		isAnonymized = true
	}
	false_ := false

	target := s
	target.isDisabled = isAnonymized
	target.isIndefinitelyDisabled = &isAnonymized
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.deleteAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) Anonymize(now time.Time) (*AccountStatus, error) {
	true_ := true
	false_ := false

	target := s
	target.isDisabled = true
	target.isIndefinitelyDisabled = &true_
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = nil
	target.isAnonymized = &true_
	target.anonymizedAt = &now

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
	target.isDisabled = true
	target.isIndefinitelyDisabled = &true_
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = &anonymizeAt
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.anonymizeAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleAnonymizationByAdmin() (*AccountStatus, error) {
	false_ := false

	target := s
	target.isDisabled = false
	target.isIndefinitelyDisabled = &false_
	target.isDeactivated = &false_
	target.disableReason = nil
	target.temporarilyDisabledFrom = nil
	target.temporarilyDisabledUntil = nil
	target.deleteAt = nil
	target.anonymizeAt = nil
	// Keep IsAnonymized unchanged.

	// FIXME(account-status): Set account_status_stale_from

	// FIXME(account-status): Allow more state transitions.
	if s.anonymizeAt == nil {
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
