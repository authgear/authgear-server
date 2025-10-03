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
	IsDisabled    bool
	IsDeactivated bool
	DisableReason *string
	DeleteAt      *time.Time
	IsAnonymized  bool
	AnonymizeAt   *time.Time
}

func (s AccountStatus) Type() AccountStatusType {
	if !s.IsDisabled {
		return AccountStatusTypeNormal
	}
	if s.DeleteAt != nil {
		if s.IsDeactivated {
			return AccountStatusTypeScheduledDeletionDeactivated
		}
		return AccountStatusTypeScheduledDeletionDisabled
	}
	if s.IsAnonymized {
		return AccountStatusTypeAnonymized
	}
	if s.AnonymizeAt != nil {
		return AccountStatusTypeScheduledAnonymizationDisabled
	}
	if s.IsDeactivated {
		return AccountStatusTypeDeactivated
	}
	return AccountStatusTypeDisabled
}

func (s AccountStatus) Check() error {
	// This method must be in sync with IsAccountStatusError.
	typ := s.Type()
	switch typ {
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
		panic(fmt.Errorf("unknown account status type: %v", typ))
	}
}

func (s AccountStatus) Reenable() (*AccountStatus, error) {
	target := AccountStatus{}
	if s.Type() == AccountStatusTypeDisabled {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) Disable(reason *string) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:    true,
		DisableReason: reason,
	}
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleDeletionByEndUser(deleteAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:    true,
		IsDeactivated: true,
		DeleteAt:      &deleteAt,
	}
	if s.Type() != AccountStatusTypeNormal {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) ScheduleDeletionByAdmin(deleteAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   true,
		IsAnonymized: s.IsAnonymized,
		DeleteAt:     &deleteAt,
	}
	if s.DeleteAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleDeletionByAdmin() (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   s.IsAnonymized,
		IsAnonymized: s.IsAnonymized,
		DeleteAt:     nil,
	}
	if s.DeleteAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) Anonymize() (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   true,
		IsAnonymized: true,
		AnonymizeAt:  s.AnonymizeAt,
	}
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleAnonymizationByAdmin(anonymizeAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:  true,
		AnonymizeAt: &anonymizeAt,
	}
	if s.AnonymizeAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleAnonymizationByAdmin() (*AccountStatus, error) {
	var target AccountStatus
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
