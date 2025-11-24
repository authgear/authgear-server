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
	accountStatusTypeDeactivated                    accountStatusType = "deactivated"
	accountStatusTypeScheduledDeletionDisabled      accountStatusType = "scheduled_deletion_disabled"
	accountStatusTypeScheduledDeletionDeactivated   accountStatusType = "scheduled_deletion_deactivated"
	accountStatusTypeScheduledAnonymizationDisabled accountStatusType = "scheduled_anonymization_disabled"

	// These 2 are special types because they may "overlap" with any other types.
	accountStatusTypeAnonymized         accountStatusType = "anonymized"
	accountStatusTypeOutsideValidPeriod accountStatusType = "outside_valid_period"
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

// type AccountStatusVariantDeactivated struct{}
//
// var _ accountStatusVariant = AccountStatusVariantDeactivated{}
//
// func (_ AccountStatusVariantDeactivated) getAccountStatusType() accountStatusType {
// 	return accountStatusTypeDeactivated
// }
//
// func (_ AccountStatusVariantDeactivated) check() error { return ErrDeactivatedUser }

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
	deleteReason             *string
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
	}.normalizeOnRead()
}

// normalizeOnRead performs necessary normalization when account status is read from the database.
func (s AccountStatusWithRefTime) normalizeOnRead() AccountStatusWithRefTime {
	// This block reads is_disabled.
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

	// The normalization applicable to account status change is also applicable to read.
	s = s.normalizeOnChange()
	return s
}

// normalizeOnChange performs necessary normalization when account status is changed.
func (s AccountStatusWithRefTime) normalizeOnChange() AccountStatusWithRefTime {
	// Remove the temporarily disabled period if it is irrelevant.
	// That is, the temporarily disabled period is in the past.
	if s.accountStatus.temporarilyDisabledFrom != nil && s.accountStatus.temporarilyDisabledUntil != nil {
		if s.refTime.After(*s.accountStatus.temporarilyDisabledFrom) && s.refTime.After(*s.accountStatus.temporarilyDisabledUntil) {
			s.accountStatus.temporarilyDisabledFrom = nil
			s.accountStatus.temporarilyDisabledUntil = nil
		}
	}
	s.accountStatus.isDisabled = s.IsDisabled()
	s.accountStatus.accountStatusStaleFrom = s.deriveAccountStatusStaleFrom()
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
			break
		}
	}

	// Step 6
	if found != nil {
		return found
	}

	// Step 7
	return nil
}

func (s AccountStatusWithRefTime) checkAllTimestampsAreValid() error {
	if s.accountStatus.accountValidFrom != nil && s.accountStatus.accountValidUntil != nil {
		if !s.accountStatus.accountValidFrom.Before(*s.accountStatus.accountValidUntil) {
			return InvalidAccountStatusTransition.New("the start timestamp of account valid period must be less than the end timestamp of the account valid period")
		}
	}

	if s.accountStatus.temporarilyDisabledFrom != nil {
		if s.accountStatus.temporarilyDisabledUntil == nil {
			return InvalidAccountStatusTransition.New("temporarily disabled period is missing the end timestamp")
		}
	}

	if s.accountStatus.temporarilyDisabledUntil != nil {
		if s.accountStatus.temporarilyDisabledFrom == nil {
			return InvalidAccountStatusTransition.New("temporarily disabled period is missing the start timestamp")
		}
	}

	if s.accountStatus.temporarilyDisabledFrom != nil && s.accountStatus.temporarilyDisabledUntil != nil {
		if !s.accountStatus.temporarilyDisabledFrom.Before(*s.accountStatus.temporarilyDisabledUntil) {
			return InvalidAccountStatusTransition.New("the start timestamp of temporarily disabled period must be less than the end timestamp of temporarily disabled period")
		}

		if s.accountStatus.accountValidFrom != nil {
			if !s.accountStatus.accountValidFrom.Before(*s.accountStatus.temporarilyDisabledFrom) {
				return InvalidAccountStatusTransition.New("the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period")
			}
		}

		if s.accountStatus.accountValidUntil != nil {
			if !s.accountStatus.temporarilyDisabledUntil.Before(*s.accountStatus.accountValidUntil) {
				return InvalidAccountStatusTransition.New("the end timestamp of temporarily disabled period must be less than the end timestamp of account valid period")
			}
		}
	}

	return nil
}

func (s AccountStatusWithRefTime) IsAnonymized() bool {
	return *s.accountStatus.isAnonymized
}

func (s AccountStatusWithRefTime) IsDisabled() bool {
	err := s.Check()
	if err != nil {
		return true
	}
	return false
}

func (s AccountStatusWithRefTime) IsDeactivated() bool {
	return *s.accountStatus.isDeactivated
}

func (s AccountStatusWithRefTime) DeleteAt() *time.Time {
	return s.accountStatus.deleteAt
}

func (s AccountStatusWithRefTime) DeleteReason() *string {
	return s.accountStatus.deleteReason
}

func (s AccountStatusWithRefTime) AnonymizeAt() *time.Time {
	return s.accountStatus.anonymizeAt
}

func (s AccountStatusWithRefTime) DisableReason() *string {
	return s.accountStatus.disableReason
}

func (s AccountStatusWithRefTime) TemporarilyDisabledFrom() *time.Time {
	return s.accountStatus.temporarilyDisabledFrom
}

func (s AccountStatusWithRefTime) TemporarilyDisabledUntil() *time.Time {
	return s.accountStatus.temporarilyDisabledUntil
}

func (s AccountStatusWithRefTime) AccountValidFrom() *time.Time {
	return s.accountStatus.accountValidFrom
}

func (s AccountStatusWithRefTime) AccountValidUntil() *time.Time {
	return s.accountStatus.accountValidUntil
}

func (s AccountStatusWithRefTime) AccountStatusStaleFrom() *time.Time {
	return s.accountStatus.accountStatusStaleFrom
}

func (s AccountStatusWithRefTime) isOutsideValidPeriod() bool {
	if s.accountStatus.accountValidFrom != nil {
		lessThanFrom := s.refTime.Before(*s.accountStatus.accountValidFrom)
		if lessThanFrom {
			return true
		}
	}

	if s.accountStatus.accountValidUntil != nil {
		equalOrGreaterThanUntil := s.refTime.Equal(*s.accountStatus.accountValidUntil) || s.refTime.After(*s.accountStatus.accountValidUntil)
		if equalOrGreaterThanUntil {
			return true
		}
	}

	return false
}

func (s AccountStatusWithRefTime) Check() error {
	if s.IsAnonymized() {
		return ErrAnonymizedUser
	}

	if s.isOutsideValidPeriod() {
		return ErrUserOutsideValidPeriod
	}

	variant := s.variant()
	err := variant.check()
	if err != nil {
		return err
	}

	return nil
}

func (s AccountStatusWithRefTime) getMostAppropriateType() accountStatusType {
	if s.IsAnonymized() {
		return accountStatusTypeAnonymized
	}
	variant := s.variant()
	type_ := variant.getAccountStatusType()
	if type_ != accountStatusTypeNormal {
		return type_
	}

	if s.isOutsideValidPeriod() {
		return accountStatusTypeOutsideValidPeriod
	}

	return accountStatusTypeNormal
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
		panic(fmt.Errorf("deactivated is not an implemented account status"))
		// return AccountStatusVariantDeactivated{}
	}

	return AccountStatusVariantDisabledIndefinitely{
		disableReason: s.accountStatus.disableReason,
	}
}

func (s AccountStatusWithRefTime) Reenable() (*AccountStatusWithRefTime, error) {
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &false_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	// case AccountStatusVariantDeactivated{}.getAccountStatusType():
	// 	return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) DisableIndefinitely(reason *string) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = reason
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) DisableTemporarily(from *time.Time, until *time.Time, reason *string) (*AccountStatusWithRefTime, error) {
	false_ := false
	target := s

	target.accountStatus.isIndefinitelyDisabled = &false_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = reason
	target.accountStatus.temporarilyDisabledFrom = from
	target.accountStatus.temporarilyDisabledUntil = until
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) SetAccountValidFrom(t *time.Time) (*AccountStatusWithRefTime, error) {
	target := s

	target.accountStatus.accountValidFrom = t
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.getMostAppropriateType())
	}

	// Account valid period can be set independently
	return &target, nil
}

func (s AccountStatusWithRefTime) SetAccountValidUntil(t *time.Time) (*AccountStatusWithRefTime, error) {
	target := s
	target.accountStatus.accountValidUntil = t
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.getMostAppropriateType())
	}

	// Account valid period can be set independently
	return &target, nil
}

func (s AccountStatusWithRefTime) SetAccountValidPeriod(from *time.Time, until *time.Time) (*AccountStatusWithRefTime, error) {
	target := s
	target.accountStatus.accountValidFrom = from
	target.accountStatus.accountValidUntil = until
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.getMostAppropriateType())
	}

	// Account valid period can be set independently
	return &target, nil
}

func (s AccountStatusWithRefTime) ScheduleDeletionByEndUser(deleteAt time.Time, reason string) (*AccountStatusWithRefTime, error) {
	true_ := true

	target := s
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &true_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = &deleteAt
	target.accountStatus.deleteReason = &reason
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) ScheduleDeletionByAdmin(deleteAt time.Time, reason string) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = &deleteAt
	target.accountStatus.deleteReason = &reason
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	// It is allowed to schedule deletion of an anonymized user.

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	default:
		return &target, nil
	}
}

func (s AccountStatusWithRefTime) UnscheduleDeletionByAdmin() (*AccountStatusWithRefTime, error) {
	isAnonymized := *s.accountStatus.isAnonymized
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &isAnonymized
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	// It is not allowed to unschedule deletion of an anonymized user.
	if s.IsAnonymized() && s.accountStatus.deleteAt == nil {
		return nil, makeTransitionError(accountStatusTypeAnonymized, accountStatusTypeAnonymized)
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) Anonymize() (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	// Keep deleteAt unchanged.
	target.accountStatus.anonymizeAt = nil
	target.accountStatus.isAnonymized = &true_
	target.accountStatus.anonymizedAt = &s.refTime
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		if target.IsAnonymized() {
			return nil, makeTransitionError(accountStatusTypeAnonymized, accountStatusTypeAnonymized)
		}
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	// case AccountStatusVariantDeactivated{}.getAccountStatusType():
	// 	return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) ScheduleAnonymizationByAdmin(anonymizeAt time.Time) (*AccountStatusWithRefTime, error) {
	true_ := true
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &true_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = &anonymizeAt
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantNormal{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledIndefinitely{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantDisabledTemporarily{}.getAccountStatusType():
		return &target, nil
	// case AccountStatusVariantDeactivated{}.getAccountStatusType():
	// 	return &target, nil
	case AccountStatusVariantScheduledDeletionByAdmin{}.getAccountStatusType():
		return &target, nil
	case AccountStatusVariantScheduledDeletionByEndUser{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func (s AccountStatusWithRefTime) UnscheduleAnonymizationByAdmin() (*AccountStatusWithRefTime, error) {
	false_ := false

	target := s
	target.accountStatus.isIndefinitelyDisabled = &false_
	target.accountStatus.isDeactivated = &false_
	target.accountStatus.disableReason = nil
	target.accountStatus.temporarilyDisabledFrom = nil
	target.accountStatus.temporarilyDisabledUntil = nil
	target.accountStatus.deleteAt = nil
	target.accountStatus.deleteReason = nil
	target.accountStatus.anonymizeAt = nil
	target = target.normalizeOnChange()

	err := target.checkAllTimestampsAreValid()
	if err != nil {
		return nil, err
	}

	if s.IsAnonymized() {
		return nil, makeTransitionError(accountStatusTypeAnonymized, target.variant().getAccountStatusType())
	}

	// Account valid period is irrelevant here.

	originalType := s.variant()
	switch originalType.getAccountStatusType() {
	case AccountStatusVariantScheduledAnonymizationByAdmin{}.getAccountStatusType():
		return &target, nil
	default:
		return nil, makeTransitionError(s.getMostAppropriateType(), target.variant().getAccountStatusType())
	}
}

func makeTransitionError(fromType accountStatusType, targetType accountStatusType) error {
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", fromType, targetType),
		map[string]interface{}{
			"from": fromType,
			"to":   targetType,
		},
	)
}

func IsAccountStatusError(err error) bool {
	// This function must be in sync with AccountStatusType.Check
	switch {
	case apierrors.IsKind(err, DisabledUser):
		return true
	// case apierrors.IsKind(err, DeactivatedUser):
	// 	return true
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
