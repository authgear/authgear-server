package user

import (
	"fmt"
	"slices"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type AccountStatusType string

const (
	AccountStatusTypeNormal                         AccountStatusType = "normal"
	AccountStatusTypeDisabled                       AccountStatusType = "disabled"
	AccountStatusTypeDisabledTemporarily            AccountStatusType = "disabled_temporarily"
	AccountStatusTypeOutsideValidPeriod             AccountStatusType = "outside_valid_period"
	AccountStatusTypeDeactivated                    AccountStatusType = "deactivated"
	AccountStatusTypeScheduledDeletionDisabled      AccountStatusType = "scheduled_deletion_disabled"
	AccountStatusTypeScheduledDeletionDeactivated   AccountStatusType = "scheduled_deletion_deactivated"
	AccountStatusTypeAnonymized                     AccountStatusType = "anonymized"
	AccountStatusTypeScheduledAnonymizationDisabled AccountStatusType = "scheduled_anonymization_disabled"
)

type AccountStatusVariant interface {
	Type() AccountStatusType
	Check() error
}

type AccountStatusVariantNormal struct{}

var _ AccountStatusVariant = AccountStatusVariantNormal{}

func (_ AccountStatusVariantNormal) Type() AccountStatusType { return AccountStatusTypeNormal }

func (_ AccountStatusVariantNormal) Check() error { return nil }

type AccountStatusVariantDisabledIndefinitely struct {
	disableReason *string
}

var _ AccountStatusVariant = AccountStatusVariantDisabledIndefinitely{}

func (_ AccountStatusVariantDisabledIndefinitely) Type() AccountStatusType {
	return AccountStatusTypeDisabled
}

func (a AccountStatusVariantDisabledIndefinitely) Check() error {
	return NewErrDisabledUser(a.disableReason)
}

type AccountStatusVariantDisabledTemporarily struct {
	disableReason            *string
	temporarilyDisabledFrom  time.Time
	temporarilyDisabledUntil time.Time
}

var _ AccountStatusVariant = AccountStatusVariantDisabledTemporarily{}

func (_ AccountStatusVariantDisabledTemporarily) Type() AccountStatusType {
	return AccountStatusTypeDisabledTemporarily
}

func (a AccountStatusVariantDisabledTemporarily) Check() error {
	return NewErrDisabledUser(a.disableReason)
}

type AccountStatusVariantOutsideValidPeriod struct {
	accountValidFrom  *time.Time
	accountValidUntil *time.Time
}

var _ AccountStatusVariant = AccountStatusVariantOutsideValidPeriod{}

func (_ AccountStatusVariantOutsideValidPeriod) Type() AccountStatusType {
	return AccountStatusTypeOutsideValidPeriod
}

func (_ AccountStatusVariantOutsideValidPeriod) Check() error { return ErrUserOutsideValidPeriod }

type AccountStatusVariantDeactivated struct{}

var _ AccountStatusVariant = AccountStatusVariantDeactivated{}

func (_ AccountStatusVariantDeactivated) Type() AccountStatusType {
	return AccountStatusTypeDeactivated
}

func (_ AccountStatusVariantDeactivated) Check() error { return ErrDeactivatedUser }

type AccountStatusVariantScheduledDeletionByAdmin struct {
	deleteAt time.Time
}

var _ AccountStatusVariant = AccountStatusVariantScheduledDeletionByAdmin{}

func (_ AccountStatusVariantScheduledDeletionByAdmin) Type() AccountStatusType {
	return AccountStatusTypeScheduledDeletionDisabled
}

func (a AccountStatusVariantScheduledDeletionByAdmin) Check() error {
	return NewErrScheduledDeletionByAdmin(a.deleteAt)
}

type AccountStatusVariantScheduledDeletionByEndUser struct {
	deleteAt time.Time
}

var _ AccountStatusVariant = AccountStatusVariantScheduledDeletionByEndUser{}

func (_ AccountStatusVariantScheduledDeletionByEndUser) Type() AccountStatusType {
	return AccountStatusTypeScheduledDeletionDeactivated
}

func (a AccountStatusVariantScheduledDeletionByEndUser) Check() error {
	return NewErrScheduledDeletionByEndUser(a.deleteAt)
}

type AccountStatusVariantAnonymized struct{}

func (_ AccountStatusVariantAnonymized) Type() AccountStatusType { return AccountStatusTypeAnonymized }

func (_ AccountStatusVariantAnonymized) Check() error {
	return ErrAnonymizedUser
}

type AccountStatusVariantScheduledAnonymizationByAdmin struct {
	anonymizeAt time.Time
}

var _ AccountStatusVariant = AccountStatusVariantScheduledAnonymizationByAdmin{}

func (_ AccountStatusVariantScheduledAnonymizationByAdmin) Type() AccountStatusType {
	return AccountStatusTypeScheduledAnonymizationDisabled
}

func (a AccountStatusVariantScheduledAnonymizationByAdmin) Check() error {
	return NewErrScheduledAnonymizationByAdmin(a.anonymizeAt)
}

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

type AccountStatusWithRefTime struct {
	accountStatus AccountStatus
	refTime       time.Time
}

func (s AccountStatus) WithRefTime(refTime time.Time) AccountStatusWithRefTime {
	return AccountStatusWithRefTime{
		accountStatus: s,
		refTime:       refTime,
	}.normalize()
}

func (s AccountStatusWithRefTime) normalize() AccountStatusWithRefTime {
	if s.accountStatus.isIndefinitelyDisabled == nil {
		if s.accountStatus.accountValidFrom == nil && s.accountStatus.accountValidUntil == nil && s.accountStatus.temporarilyDisabledFrom == nil && s.accountStatus.temporarilyDisabledUntil == nil {
			// It is safe to read isDisabled here because the columns
			// that can affect account_status_stale_from are nulls.
			disabled := s.accountStatus.isDisabled
			s.accountStatus.isIndefinitelyDisabled = &disabled
		} else {
			panic(fmt.Errorf("is_indefinitely_disabled should have been patched"))
		}
	}

	if s.accountStatus.isDeactivated == nil {
		false_ := false
		s.accountStatus.isDeactivated = &false_
	}

	if s.accountStatus.isAnonymized == nil {
		false_ := false
		s.accountStatus.isAnonymized = &false_
	}

	return s
}

func (s AccountStatusWithRefTime) deriveAccountStatusStaleFrom() *time.Time {
	// Step 1
	var times []time.Time
	if t := s.accountStatus.accountValidFrom; t != nil {
		times = append(times, *t)
	}
	if t := s.accountStatus.accountValidUntil; t != nil {
		times = append(times, *t)
	}
	if t := s.accountStatus.temporarilyDisabledFrom; t != nil {
		times = append(times, *t)
	}
	if t := s.accountStatus.temporarilyDisabledUntil; t != nil {
		times = append(times, *t)
	}

	// Step 2
	slices.SortFunc(times, func(a time.Time, b time.Time) int {
		return a.Compare(b)
	})

	// Step 3
	if len(times) == 0 {
		return nil
	}

	// Step 5
	var found *time.Time
	for _, t := range times {
		if t.After(s.refTime) {
			t := t
			found = &t
		}
	}

	// Step 6
	if found != nil {
		return found
	}

	// Step 7
	return nil
}

func (s AccountStatusWithRefTime) Variant() AccountStatusVariant {
	// This method does not read is_disabled,
	// thus account_status_stale_from is irrelevant here.

	if !*s.accountStatus.isIndefinitelyDisabled {
		if s.accountStatus.temporarilyDisabledFrom != nil && s.accountStatus.temporarilyDisabledUntil != nil {
			equalOrGreaterThanFrom := s.refTime.Equal(*s.accountStatus.temporarilyDisabledFrom) || s.refTime.After(*s.accountStatus.temporarilyDisabledFrom)
			lessThanUntil := s.refTime.Before(*s.accountStatus.temporarilyDisabledUntil)

			if equalOrGreaterThanFrom && lessThanUntil {
				return AccountStatusVariantDisabledTemporarily{
					disableReason:            s.accountStatus.disableReason,
					temporarilyDisabledFrom:  *s.accountStatus.temporarilyDisabledFrom,
					temporarilyDisabledUntil: *s.accountStatus.temporarilyDisabledUntil,
				}
			}
		}

		if s.accountStatus.accountValidFrom != nil {
			lessThanFrom := s.refTime.Before(*s.accountStatus.accountValidFrom)
			if lessThanFrom {
				return AccountStatusVariantOutsideValidPeriod{
					accountValidFrom:  s.accountStatus.accountValidFrom,
					accountValidUntil: s.accountStatus.accountValidUntil,
				}
			}
		}

		if s.accountStatus.accountValidUntil != nil {
			equalOrGreaterThanUntil := s.refTime.Equal(*s.accountStatus.accountValidUntil) || s.refTime.After(*s.accountStatus.accountValidUntil)
			if equalOrGreaterThanUntil {
				return AccountStatusVariantOutsideValidPeriod{
					accountValidFrom:  s.accountStatus.accountValidFrom,
					accountValidUntil: s.accountStatus.accountValidUntil,
				}
			}
		}

		return AccountStatusVariantNormal{}
	}

	if s.accountStatus.deleteAt != nil {
		if *s.accountStatus.isDeactivated {
			return AccountStatusVariantScheduledDeletionByEndUser{
				deleteAt: *s.accountStatus.deleteAt,
			}
		}
		return AccountStatusVariantScheduledDeletionByAdmin{
			deleteAt: *s.accountStatus.deleteAt,
		}
	}
	if *s.accountStatus.isAnonymized {
		return AccountStatusVariantAnonymized{}
	}
	if s.accountStatus.anonymizeAt != nil {
		return AccountStatusVariantScheduledAnonymizationByAdmin{
			anonymizeAt: *s.accountStatus.anonymizeAt,
		}
	}
	if *s.accountStatus.isDeactivated {
		return AccountStatusVariantDeactivated{}
	}

	return AccountStatusVariantDisabledIndefinitely{
		disableReason: s.accountStatus.disableReason,
	}
}

func (s AccountStatusWithRefTime) Reenable() (*AccountStatusWithRefTime, error) {
	false_ := false

	target := s
	target.accountStatus.isDisabled = false
	target.accountStatus.isIndefinitelyDisabled = &false_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.isAnonymized = &false_
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantDisabledIndefinitely{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.Type():
		return &target, nil
	case AccountStatusVariantDeactivated{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) Disable(reason *string) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = reason
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.isAnonymized = &false_
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantNormal{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) ScheduleDeletionByEndUser(deleteAt time.Time) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &true_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = &deleteAt
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.isAnonymized = &false_
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantNormal{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) ScheduleDeletionByAdmin(deleteAt time.Time) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = &deleteAt
	target.accountStatus.anonymizeAt = nil
	// Keep IsAnonymized unchanged.
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.Type():
		return nil, makeTransitionError(originalType, target.Variant())
	case AccountStatusVariantScheduledDeletionByEndUser{}.Type():
		return nil, makeTransitionError(originalType, target.Variant())
	default:
		return &target, nil
	}
}

func (s AccountStatusWithRefTime) UnscheduleDeletionByAdmin() (*AccountStatusWithRefTime, error) {
	isAnonymized := *s.accountStatus.isAnonymized
	false_ := false

	target := s
	target.accountStatus.isDisabled = isAnonymized
	target.accountStatus.isIndefinitelyDisabled = &isAnonymized
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.anonymizeAt = nil
	// Keep IsAnonymized unchanged.
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) Anonymize() (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	// Keep deleteAt unchanged.
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.isAnonymized = &true_
	target.accountStatus.anonymizedAt = &s.refTime
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantNormal{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.Type():
		return &target, nil
	case AccountStatusVariantOutsideValidPeriod{}.Type():
		return &target, nil
	case AccountStatusVariantDeactivated{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) ScheduleAnonymizationByAdmin(anonymizeAt time.Time) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.anonymizeAt = &anonymizeAt
	// Keep IsAnonymized unchanged.
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantNormal{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.Type():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.Type():
		return &target, nil
	case AccountStatusVariantOutsideValidPeriod{}.Type():
		return &target, nil
	case AccountStatusVariantDeactivated{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.Type():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func (s AccountStatusWithRefTime) UnscheduleAnonymizationByAdmin() (*AccountStatusWithRefTime, error) {
	false_ := false

	target := s
	target.accountStatus.isDisabled = false
	target.accountStatus.isIndefinitelyDisabled = &false_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.anonymizeAt = nil
	// Keep IsAnonymized unchanged.
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	originalType := s.Variant()
	switch originalType.Type() {
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.Type():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.Variant())
	}
}

func makeTransitionError(fromType AccountStatusVariant, targetType AccountStatusVariant) error {
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", fromType.Type(), targetType.Type()),
		map[string]interface{}{
			"from": fromType.Type(),
			"to":   targetType.Type(),
		},
	)
}

func IsAccountStatusError(err error) bool {
	// This function must be in sync with AccountStatusType.Check
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
	case apierrors.IsKind(err, UserOutsideValidPeriod):
		return true
	default:
		return false
	}
}
