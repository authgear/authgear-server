package user

import (
	"fmt"
	"slices"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type accountStatusType string

const (
	accountStatusTypeNormal                         accountStatusType = "normal"
	accountStatusTypeDisabled                       accountStatusType = "disabled"
	accountStatusTypeDisabledTemporarily            accountStatusType = "disabled_temporarily"
	accountStatusTypeOutsideValidPeriod             accountStatusType = "outside_valid_period"
	accountStatusTypeDeactivated                    accountStatusType = "deactivated"
	accountStatusTypeScheduledDeletionDisabled      accountStatusType = "scheduled_deletion_disabled"
	accountStatusTypeScheduledDeletionDeactivated   accountStatusType = "scheduled_deletion_deactivated"
	accountStatusTypeScheduledAnonymizationDisabled accountStatusType = "scheduled_anonymization_disabled"
)

type accountStatusVariant interface {
	getAccountStatusType() accountStatusType
	check() error
}

type AccountStatusVariantNormal struct{}

var _ accountStatusVariant = AccountStatusVariantNormal{}

func (_ AccountStatusVariantNormal) getAccountStatusType() accountStatusType {
	return accountStatusTypeNormal
}

func (_ AccountStatusVariantNormal) check() error { return nil }

type AccountStatusVariantDisabledIndefinitely struct {
	disableReason *string
}

var _ accountStatusVariant = AccountStatusVariantDisabledIndefinitely{}

func (_ AccountStatusVariantDisabledIndefinitely) getAccountStatusType() accountStatusType {
	return accountStatusTypeDisabled
}

func (a AccountStatusVariantDisabledIndefinitely) check() error {
	return NewErrDisabledUser(a.disableReason)
}

type AccountStatusVariantDisabledTemporarily struct {
	disableReason            *string
	temporarilyDisabledFrom  time.Time
	temporarilyDisabledUntil time.Time
}

var _ accountStatusVariant = AccountStatusVariantDisabledTemporarily{}

func (_ AccountStatusVariantDisabledTemporarily) getAccountStatusType() accountStatusType {
	return accountStatusTypeDisabledTemporarily
}

func (a AccountStatusVariantDisabledTemporarily) check() error {
	return NewErrDisabledUser(a.disableReason)
}

type AccountStatusVariantOutsideValidPeriod struct {
	accountValidFrom  *time.Time
	accountValidUntil *time.Time
}

var _ accountStatusVariant = AccountStatusVariantOutsideValidPeriod{}

func (_ AccountStatusVariantOutsideValidPeriod) getAccountStatusType() accountStatusType {
	return accountStatusTypeOutsideValidPeriod
}

func (_ AccountStatusVariantOutsideValidPeriod) check() error { return ErrUserOutsideValidPeriod }

type AccountStatusVariantDeactivated struct{}

var _ accountStatusVariant = AccountStatusVariantDeactivated{}

func (_ AccountStatusVariantDeactivated) getAccountStatusType() accountStatusType {
	return accountStatusTypeDeactivated
}

func (_ AccountStatusVariantDeactivated) check() error { return ErrDeactivatedUser }

type AccountStatusVariantScheduledDeletionByAdmin struct {
	deleteAt time.Time
}

var _ accountStatusVariant = AccountStatusVariantScheduledDeletionByAdmin{}

func (_ AccountStatusVariantScheduledDeletionByAdmin) getAccountStatusType() accountStatusType {
	return accountStatusTypeScheduledDeletionDisabled
}

func (a AccountStatusVariantScheduledDeletionByAdmin) check() error {
	return NewErrScheduledDeletionByAdmin(a.deleteAt)
}

type AccountStatusVariantScheduledDeletionByEndUser struct {
	deleteAt time.Time
}

var _ accountStatusVariant = AccountStatusVariantScheduledDeletionByEndUser{}

func (_ AccountStatusVariantScheduledDeletionByEndUser) getAccountStatusType() accountStatusType {
	return accountStatusTypeScheduledDeletionDeactivated
}

func (a AccountStatusVariantScheduledDeletionByEndUser) check() error {
	return NewErrScheduledDeletionByEndUser(a.deleteAt)
}

type AccountStatusVariantScheduledAnonymizationByAdmin struct {
	anonymizeAt time.Time
}

var _ accountStatusVariant = AccountStatusVariantScheduledAnonymizationByAdmin{}

func (_ AccountStatusVariantScheduledAnonymizationByAdmin) getAccountStatusType() accountStatusType {
	return accountStatusTypeScheduledAnonymizationDisabled
}

func (a AccountStatusVariantScheduledAnonymizationByAdmin) check() error {
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

func (s AccountStatusWithRefTime) IsAnonymized() bool {
	return *s.accountStatus.isAnonymized
}

func (s AccountStatusWithRefTime) Check() error {
	if s.IsAnonymized() {
		return ErrAnonymizedUser
	}

	variant := s.variant()
	return variant.check()
}

func (s AccountStatusWithRefTime) variant() accountStatusVariant {
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	if s.IsAnonymized() {
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDeactivated{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
	}
}

func (s AccountStatusWithRefTime) DisableIndefinitely(reason *string) (*AccountStatusWithRefTime, error) {
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	if s.IsAnonymized() {
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
	}
}

func (s AccountStatusWithRefTime) ScheduleDeletionByEndUser(deleteAt time.Time) (*AccountStatusWithRefTime, error) {
	true_ := true

	target := s
	target.accountStatus.isDisabled = true
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &true_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = &deleteAt
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	if s.IsAnonymized() {
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	// It is allowed to schedule deletion of an anonymized user.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return nil, makeTransitionError(originalType, target.variant())
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return nil, makeTransitionError(originalType, target.variant())
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	// It is allowed to unschedule deletion of an anonymized user.
	if s.IsAnonymized() && s.accountStatus.deleteAt == nil {
		return nil, makeTransitionErrorFromAnonymizedToAnonymized()
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
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

	if s.IsAnonymized() {
		if target.IsAnonymized() {
			return nil, makeTransitionErrorFromAnonymizedToAnonymized()
		}
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantOutsideValidPeriod{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDeactivated{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	if s.IsAnonymized() {
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantOutsideValidPeriod{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDeactivated{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
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
	target.accountStatus.accountStatusStaleFrom = target.deriveAccountStatusStaleFrom()

	if s.IsAnonymized() {
		return nil, makeTransitionErrorFromAnonymized(target.variant())
	}

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(originalType, target.variant())
	}
}

func makeTransitionError(fromType accountStatusVariant, targetType accountStatusVariant) error {
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", fromType.getAccountStatusType(), targetType.getAccountStatusType()),
		map[string]interface{}{
			"from": fromType.getAccountStatusType(),
			"to":   targetType.getAccountStatusType(),
		},
	)
}

func makeTransitionErrorFromAnonymized(targetType accountStatusVariant) error {
	label := "anonymized"
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", label, targetType.getAccountStatusType()),
		map[string]interface{}{
			"from": label,
			"to":   targetType.getAccountStatusType(),
		},
	)
}

func makeTransitionErrorFromAnonymizedToAnonymized() error {
	label := "anonymized"
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", label, label),
		map[string]interface{}{
			"from": label,
			"to":   label,
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
