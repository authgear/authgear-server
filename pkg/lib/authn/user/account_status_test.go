package user

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type accountStatusStateTransitionTest struct {
	Reenable                       string
	DisableIndefinitely            string
	DisableTemporarily_Now         string
	DisableTemporarily_Future      string
	SetAccountValidPeriod_Inside   string
	SetAccountValidPeriod_Outside  string
	ScheduleDeletionByEndUser      string
	ScheduleDeletionByAdmin        string
	UnscheduleDeletionByAdmin      string
	Anonymize                      string
	ScheduleAnonymizationByAdmin   string
	UnscheduleAnonymizationByAdmin string
}

func TestAccountStatusNormalization(t *testing.T) {
	Convey("AccountStatus normalization", t, func() {
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)

		Convey("isIndefinitelyDisabled is normalized upon construction", func() {
			legacyDisabled := AccountStatus{
				isDisabled: true,
			}.WithRefTime(now)
			So(legacyDisabled.accountStatus.isIndefinitelyDisabled, ShouldNotBeNil)
			So(*legacyDisabled.accountStatus.isIndefinitelyDisabled, ShouldEqual, true)

			legacyNormal := AccountStatus{}.WithRefTime(now)
			So(legacyNormal.accountStatus.isIndefinitelyDisabled, ShouldNotBeNil)
			So(*legacyNormal.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
		})

		Convey("isDeactivated is never nil", func() {
			normal := AccountStatus{}.WithRefTime(now)

			So(normal.accountStatus.isDeactivated, ShouldNotBeNil)
			So(*normal.accountStatus.isDeactivated, ShouldEqual, false)
		})

		Convey("isAnonymized is never nil", func() {
			normal := AccountStatus{}.WithRefTime(now)

			So(normal.accountStatus.isAnonymized, ShouldNotBeNil)
			So(*normal.accountStatus.isAnonymized, ShouldEqual, false)
		})
	})
}

//nolint:gocognit // Further splitting this test function does not improve readability.
func TestAccountStatusStateTransition(t *testing.T) {
	Convey("AccountStatus state transition", t, func() {
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		deleteAt := now
		anonymizeAt := now

		temporarilyDisabledFrom := now
		temporarilyDisabledUntil := now.Add(time.Hour * 24)

		accountValidFrom_outsideValidPeriod := now.Add(time.Hour * 24)
		accountValidUntil_outsideValidPeriod := now.Add(time.Hour * 24 * 2)

		accountValidFrom_insideValidPeriod := now
		accountValidUntil_insideValidPeriod := now.Add(time.Hour * 24)

		temporarilyDisabledFromInFuture := now.Add(time.Hour * 24 * 1)
		temporarilyDisabledUntilInFuture := now.Add(time.Hour * 24 * 2)

		temporarilyDisabledFromInPast := now.Add(-time.Hour * 24 * 2)
		temporarilyDisabledUntilInPast := now.Add(-time.Hour * 24 * 1)

		testStateTransition := func(status AccountStatusWithRefTime, testCase accountStatusStateTransitionTest) {
			var err error

			_, err = status.Reenable()
			if testCase.Reenable == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.Reenable)
			}

			_, err = status.DisableIndefinitely(nil)
			if testCase.DisableIndefinitely == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.DisableIndefinitely)
			}

			_, err = status.DisableTemporarily(&temporarilyDisabledFrom, &temporarilyDisabledUntil, nil)
			if testCase.DisableTemporarily_Now == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.DisableTemporarily_Now)
			}

			_, err = status.DisableTemporarily(&temporarilyDisabledFromInFuture, &temporarilyDisabledUntilInFuture, nil)
			if testCase.DisableTemporarily_Future == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.DisableTemporarily_Future)
			}

			_, err = status.SetAccountValidPeriod(&accountValidFrom_insideValidPeriod, &accountValidUntil_insideValidPeriod)
			if testCase.SetAccountValidPeriod_Inside == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.SetAccountValidPeriod_Inside)
			}

			_, err = status.SetAccountValidPeriod(&accountValidFrom_outsideValidPeriod, &accountValidUntil_outsideValidPeriod)
			if testCase.SetAccountValidPeriod_Outside == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.SetAccountValidPeriod_Outside)
			}

			_, err = status.ScheduleDeletionByEndUser(deleteAt)
			if testCase.ScheduleDeletionByEndUser == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.ScheduleDeletionByEndUser)
			}

			_, err = status.ScheduleDeletionByAdmin(deleteAt)
			if testCase.ScheduleDeletionByAdmin == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.ScheduleDeletionByAdmin)
			}

			_, err = status.UnscheduleDeletionByAdmin()
			if testCase.UnscheduleDeletionByAdmin == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.UnscheduleDeletionByAdmin)
			}

			_, err = status.Anonymize()
			if testCase.Anonymize == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.Anonymize)
			}

			_, err = status.ScheduleAnonymizationByAdmin(deleteAt)
			if testCase.ScheduleAnonymizationByAdmin == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.ScheduleAnonymizationByAdmin)
			}

			_, err = status.UnscheduleAnonymizationByAdmin()
			if testCase.UnscheduleAnonymizationByAdmin == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.UnscheduleAnonymizationByAdmin)
			}
		}

		Convey("state transition from normal", func() {
			normal := AccountStatus{}.WithRefTime(now)
			testStateTransition(normal, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: normal -> normal",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: normal -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: normal -> normal",
			})
		})

		Convey("state transition from disabled", func() {
			true_ := true
			disabled := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
			}.WithRefTime(now)

			testStateTransition(disabled, accountStatusStateTransitionTest{
				Reenable:                       "",
				DisableIndefinitely:            "invalid account status transition: disabled -> disabled",
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "invalid account status transition: disabled -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: disabled -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: disabled -> normal",
			})
		})

		Convey("state transition from disabled_temporarily", func() {
			false_ := false
			disabledTemporarily := AccountStatus{
				isDisabled:               true,
				isIndefinitelyDisabled:   &false_,
				temporarilyDisabledFrom:  &temporarilyDisabledFrom,
				temporarilyDisabledUntil: &temporarilyDisabledUntil,
			}.WithRefTime(now)

			testStateTransition(disabledTemporarily, accountStatusStateTransitionTest{
				Reenable:                       "",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
				SetAccountValidPeriod_Inside:   "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				SetAccountValidPeriod_Outside:  "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				ScheduleDeletionByEndUser:      "invalid account status transition: disabled_temporarily -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: disabled_temporarily -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: disabled_temporarily -> normal",
			})
		})

		Convey("state transition from disabled_temporarily (future) === normal", func() {
			false_ := false
			disabledTemporarily := AccountStatus{
				isDisabled:               false,
				isIndefinitelyDisabled:   &false_,
				temporarilyDisabledFrom:  &temporarilyDisabledFromInFuture,
				temporarilyDisabledUntil: &temporarilyDisabledUntilInFuture,
			}.WithRefTime(now)

			testStateTransition(disabledTemporarily, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: normal -> normal",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
				SetAccountValidPeriod_Inside:   "the end timestamp of temporarily disabled period must be less than the end timestamp of account valid period",
				SetAccountValidPeriod_Outside:  "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				ScheduleDeletionByEndUser:      "",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: normal -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: normal -> normal",
			})
		})

		Convey("state transition from disabled_temporarily (past) === normal", func() {
			false_ := false
			disabledTemporarily := AccountStatus{
				isDisabled:               false,
				isIndefinitelyDisabled:   &false_,
				temporarilyDisabledFrom:  &temporarilyDisabledFromInPast,
				temporarilyDisabledUntil: &temporarilyDisabledUntilInPast,
			}.WithRefTime(now)

			testStateTransition(disabledTemporarily, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: normal -> normal",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: normal -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: normal -> normal",
			})
		})

		Convey("state transition from outside_valid_period", func() {
			false_ := false
			outsideValidPeriod := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &false_,
				accountValidFrom:       &accountValidFrom_outsideValidPeriod,
				accountValidUntil:      &accountValidUntil_outsideValidPeriod,
			}.WithRefTime(now)

			testStateTransition(outsideValidPeriod, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: outside_valid_period -> normal",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				DisableTemporarily_Future:      "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: outside_valid_period -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: outside_valid_period -> normal",
			})
		})

		Convey("state transition from outside_valid_period (within valid period)", func() {
			false_ := false
			outsideValidPeriod := AccountStatus{
				isDisabled:             false,
				isIndefinitelyDisabled: &false_,
				accountValidFrom:       &accountValidFrom_insideValidPeriod,
				accountValidUntil:      &accountValidUntil_insideValidPeriod,
			}.WithRefTime(now)

			testStateTransition(outsideValidPeriod, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: normal -> normal",
				DisableIndefinitely:            "",
				DisableTemporarily_Now:         "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period",
				DisableTemporarily_Future:      "the end timestamp of temporarily disabled period must be less than the end timestamp of account valid period",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: normal -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: normal -> normal",
			})
		})

		Convey("state transition from scheduled_deletion_disabled", func() {
			true_ := true
			scheduledDeletion := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				deleteAt:               &deleteAt,
			}.WithRefTime(now)

			testStateTransition(scheduledDeletion, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: scheduled_deletion_disabled -> normal",
				DisableIndefinitely:            "invalid account status transition: scheduled_deletion_disabled -> disabled",
				DisableTemporarily_Now:         "invalid account status transition: scheduled_deletion_disabled -> disabled_temporarily",
				DisableTemporarily_Future:      "invalid account status transition: scheduled_deletion_disabled -> normal",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "invalid account status transition: scheduled_deletion_disabled -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "invalid account status transition: scheduled_deletion_disabled -> scheduled_deletion_disabled",
				UnscheduleDeletionByAdmin:      "",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: scheduled_deletion_disabled -> normal",
			})
		})

		Convey("state transition from scheduled_deletion_deactivated", func() {
			true_ := true
			scheduledDeletion := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isDeactivated:          &true_,
				deleteAt:               &deleteAt,
			}.WithRefTime(now)

			testStateTransition(scheduledDeletion, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: scheduled_deletion_deactivated -> normal",
				DisableIndefinitely:            "invalid account status transition: scheduled_deletion_deactivated -> disabled",
				DisableTemporarily_Now:         "invalid account status transition: scheduled_deletion_deactivated -> disabled_temporarily",
				DisableTemporarily_Future:      "invalid account status transition: scheduled_deletion_deactivated -> normal",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "invalid account status transition: scheduled_deletion_deactivated -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "invalid account status transition: scheduled_deletion_deactivated -> scheduled_deletion_disabled",
				UnscheduleDeletionByAdmin:      "",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: scheduled_deletion_deactivated -> normal",
			})
		})

		Convey("state transition from anonymized", func() {
			true_ := true
			anonymized := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				anonymizeAt:            &anonymizeAt,
			}.WithRefTime(now)

			testStateTransition(anonymized, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: anonymized -> normal",
				DisableIndefinitely:            "invalid account status transition: anonymized -> disabled",
				DisableTemporarily_Now:         "invalid account status transition: anonymized -> disabled_temporarily",
				DisableTemporarily_Future:      "invalid account status transition: anonymized -> normal",
				SetAccountValidPeriod_Inside:   "invalid account status transition: anonymized -> anonymized",
				SetAccountValidPeriod_Outside:  "invalid account status transition: anonymized -> anonymized",
				ScheduleDeletionByEndUser:      "invalid account status transition: anonymized -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: anonymized -> anonymized",
				Anonymize:                      "invalid account status transition: anonymized -> anonymized",
				ScheduleAnonymizationByAdmin:   "invalid account status transition: anonymized -> scheduled_anonymization_disabled",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: anonymized -> normal",
			})
		})

		Convey("state transition from scheduled_anonymization_disabled", func() {
			true_ := true
			scheduledAnonymization := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				anonymizeAt:            &anonymizeAt,
			}.WithRefTime(now)

			testStateTransition(scheduledAnonymization, accountStatusStateTransitionTest{
				Reenable:                       "invalid account status transition: scheduled_anonymization_disabled -> normal",
				DisableIndefinitely:            "invalid account status transition: scheduled_anonymization_disabled -> disabled",
				DisableTemporarily_Now:         "invalid account status transition: scheduled_anonymization_disabled -> disabled_temporarily",
				DisableTemporarily_Future:      "invalid account status transition: scheduled_anonymization_disabled -> normal",
				SetAccountValidPeriod_Inside:   "",
				SetAccountValidPeriod_Outside:  "",
				ScheduleDeletionByEndUser:      "invalid account status transition: scheduled_anonymization_disabled -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: scheduled_anonymization_disabled -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "invalid account status transition: scheduled_anonymization_disabled -> scheduled_anonymization_disabled",
				UnscheduleAnonymizationByAdmin: "",
			})
		})
	})
}

func TestAccountStatusAnonymizd(t *testing.T) {
	Convey("AccountStatus anonymized", t, func() {
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		deleteAt := now
		anonymizeAt := now

		Convey("anonymized -> scheduled_deletion_disabled -> anonymized", func() {
			true_ := true
			anonymized := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				anonymizeAt:            &anonymizeAt,
			}.WithRefTime(now)

			state1, err := anonymized.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(state1.IsAnonymized(), ShouldEqual, true)
			So(state1.variant().getAccountStatusType(), ShouldEqual, accountStatusTypeScheduledDeletionDisabled)

			state2, err := state1.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(state2.IsAnonymized(), ShouldEqual, true)
		})
	})
}

func TestAccountStatusTemporarilyDisabled(t *testing.T) {
	Convey("AccountStatus TemporarilyDisabled", t, func() {
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		temporarilyDisabledFrom := now
		temporarilyDisabledUntil := now.Add(time.Hour * 24)

		temporarilyDisabledFromInFuture := now.Add(time.Hour * 24 * 1)
		temporarilyDisabledUntilInFuture := now.Add(time.Hour * 24 * 2)

		Convey("isDisabled is true if now is within temporarily disabled period", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.DisableTemporarily(&temporarilyDisabledFrom, &temporarilyDisabledUntil, nil)
			So(err, ShouldBeNil)
			So(state1.IsDisabled(), ShouldEqual, true)
			So(state1.accountStatus.isDisabled, ShouldEqual, true)
			So(*state1.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isDeactivated, ShouldEqual, false)
		})

		Convey("isDisabled is false if now is NOT within temporarily disabled period", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.DisableTemporarily(&temporarilyDisabledFromInFuture, &temporarilyDisabledUntilInFuture, nil)
			So(err, ShouldBeNil)
			So(state1.IsDisabled(), ShouldEqual, false)
			So(state1.accountStatus.isDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isDeactivated, ShouldEqual, false)
		})
	})
}

func TestAccountStatusAccountValidPeriod(t *testing.T) {
	Convey("AccountStatus account valid period", t, func() {
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		deleteAt := now

		accountValidFrom_outsideValidPeriod := now.Add(time.Hour * 24)
		accountValidUntil_outsideValidPeriod := now.Add(time.Hour * 24 * 2)

		accountValidFrom_insideValidPeriod := now
		accountValidUntil_insideValidPeriod := now.Add(time.Hour * 24)

		Convey("isDisabled is false if now is within account valid period", func() {
			false_ := false
			accountStatus := AccountStatus{
				isIndefinitelyDisabled: &false_,
				accountValidFrom:       &accountValidFrom_insideValidPeriod,
				accountValidUntil:      &accountValidUntil_insideValidPeriod,
			}.WithRefTime(now)

			So(accountStatus.IsDisabled(), ShouldEqual, false)
			So(accountStatus.getMostAppropriateType(), ShouldEqual, accountStatusTypeNormal)
			So(accountStatus.accountStatus.isDisabled, ShouldEqual, false)
			So(*accountStatus.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*accountStatus.accountStatus.isDeactivated, ShouldEqual, false)
		})

		Convey("isDisabled is true if now is NOT within account valid period", func() {
			false_ := false
			accountStatus := AccountStatus{
				isIndefinitelyDisabled: &false_,
				accountValidFrom:       &accountValidFrom_outsideValidPeriod,
				accountValidUntil:      &accountValidUntil_outsideValidPeriod,
			}.WithRefTime(now)

			So(accountStatus.IsDisabled(), ShouldEqual, true)
			So(accountStatus.getMostAppropriateType(), ShouldEqual, accountStatusTypeOutsideValidPeriod)
			So(accountStatus.accountStatus.isDisabled, ShouldEqual, true)
			So(*accountStatus.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*accountStatus.accountStatus.isDeactivated, ShouldEqual, false)
		})

		Convey("normal -> outside_valid_period -> normal", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.SetAccountValidPeriod(&accountValidFrom_outsideValidPeriod, &accountValidUntil_outsideValidPeriod)
			So(err, ShouldBeNil)
			So(state1.Check(), ShouldNotBeNil)
			So(state1.IsDisabled(), ShouldEqual, true)
			So(state1.accountStatus.isDisabled, ShouldEqual, true)

			state2, err := state1.SetAccountValidPeriod(nil, nil)
			So(err, ShouldBeNil)
			So(state2.Check(), ShouldBeNil)
			So(state2.IsDisabled(), ShouldEqual, false)
			So(state2.accountStatus.isDisabled, ShouldEqual, false)
		})

		Convey("normal -> outside_valid_period -> scheduled_deletion_disabled -> outside_valid_period", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.SetAccountValidPeriod(&accountValidFrom_outsideValidPeriod, &accountValidUntil_outsideValidPeriod)
			So(err, ShouldBeNil)
			So(state1.Check(), ShouldNotBeNil)
			So(state1.Check(), ShouldBeError, "user is outside valid period")
			So(state1.IsDisabled(), ShouldEqual, true)
			So(state1.accountStatus.isDisabled, ShouldEqual, true)

			state2, err := state1.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(state2.Check(), ShouldNotBeNil)
			So(state2.Check(), ShouldBeError, "user is outside valid period")
			So(state2.IsDisabled(), ShouldEqual, true)
			So(state2.accountStatus.isDisabled, ShouldEqual, true)
			So(state2.variant().getAccountStatusType(), ShouldEqual, accountStatusTypeScheduledDeletionDisabled)

			state3, err := state2.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(state3.Check(), ShouldNotBeNil)
			So(state3.Check(), ShouldBeError, "user is outside valid period")
			So(state3.IsDisabled(), ShouldEqual, true)
			So(state3.accountStatus.isDisabled, ShouldEqual, true)
		})
	})
}

func TestAccountStatusAccountStatusStaleFrom(t *testing.T) {
	Convey("account_status_stale_from is accurate", t, func() {
		t0 := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		t1 := time.Date(2006, 1, 2, 3, 4, 5, 6+1, time.UTC)
		t2 := time.Date(2006, 1, 2, 3, 4, 5, 6+2, time.UTC)
		t3 := time.Date(2006, 1, 2, 3, 4, 5, 6+3, time.UTC)
		t4 := time.Date(2006, 1, 2, 3, 4, 5, 6+4, time.UTC)

		false_ := false
		status := AccountStatus{
			isIndefinitelyDisabled:   &false_,
			accountValidFrom:         &t0,
			accountValidUntil:        &t4,
			temporarilyDisabledFrom:  &t1,
			temporarilyDisabledUntil: &t3,
		}.WithRefTime(t2)
		So(status.accountStatus.accountStatusStaleFrom, ShouldNotBeNil)
		So(*status.accountStatus.accountStatusStaleFrom, ShouldEqual, t3)
		So(status.IsDisabled(), ShouldEqual, true)
		So(status.accountStatus.isDisabled, ShouldEqual, true)
		So(status.Check(), ShouldBeError, "user is disabled")

		status = AccountStatus{
			isIndefinitelyDisabled:   &false_,
			accountValidFrom:         &t0,
			accountValidUntil:        &t4,
			temporarilyDisabledFrom:  &t2,
			temporarilyDisabledUntil: &t3,
		}.WithRefTime(t1)
		So(status.accountStatus.accountStatusStaleFrom, ShouldNotBeNil)
		So(*status.accountStatus.accountStatusStaleFrom, ShouldEqual, t2)
		So(status.IsDisabled(), ShouldEqual, false)
		So(status.accountStatus.isDisabled, ShouldEqual, false)
		So(status.Check(), ShouldBeNil)

		status = AccountStatus{
			isIndefinitelyDisabled:   &false_,
			accountValidFrom:         &t1,
			accountValidUntil:        &t4,
			temporarilyDisabledFrom:  &t2,
			temporarilyDisabledUntil: &t3,
		}.WithRefTime(t0)
		So(status.accountStatus.accountStatusStaleFrom, ShouldNotBeNil)
		So(*status.accountStatus.accountStatusStaleFrom, ShouldEqual, t1)
		So(status.IsDisabled(), ShouldEqual, true)
		So(status.accountStatus.isDisabled, ShouldEqual, true)
		So(status.Check(), ShouldBeError, "user is outside valid period")

		status = AccountStatus{
			isIndefinitelyDisabled:   &false_,
			accountValidFrom:         &t0,
			accountValidUntil:        &t4,
			temporarilyDisabledFrom:  &t1,
			temporarilyDisabledUntil: &t2,
		}.WithRefTime(t3)
		So(status.accountStatus.accountStatusStaleFrom, ShouldNotBeNil)
		So(*status.accountStatus.accountStatusStaleFrom, ShouldEqual, t4)
		So(status.IsDisabled(), ShouldEqual, false)
		So(status.accountStatus.isDisabled, ShouldEqual, false)
		So(status.Check(), ShouldBeNil)

		status = AccountStatus{
			isIndefinitelyDisabled:   &false_,
			accountValidFrom:         &t0,
			accountValidUntil:        &t3,
			temporarilyDisabledFrom:  &t1,
			temporarilyDisabledUntil: &t2,
		}.WithRefTime(t4)
		So(status.accountStatus.accountStatusStaleFrom, ShouldBeNil)
		So(status.IsDisabled(), ShouldEqual, true)
		So(status.accountStatus.isDisabled, ShouldEqual, true)
		So(status.Check(), ShouldBeError, "user is outside valid period")
	})
}

func TestAccountStatusTimestampsValidation(t *testing.T) {
	Convey("temporarily disabled period and account valid period timestamps are validated", t, func() {
		t0 := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		t1 := time.Date(2006, 1, 2, 3, 4, 5, 6+1, time.UTC)
		t2 := time.Date(2006, 1, 2, 3, 4, 5, 6+2, time.UTC)
		t3 := time.Date(2006, 1, 2, 3, 4, 5, 6+3, time.UTC)
		t4 := time.Date(2006, 1, 2, 3, 4, 5, 6+4, time.UTC)

		normal := AccountStatus{}.WithRefTime(t0)

		Convey("temporarily disabled period must have the start timestamp", func() {
			_, err := normal.DisableTemporarily(nil, &t1, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "temporarily disabled period is missing the start timestamp")
		})

		Convey("temporarily disabled period must have the end timestamp", func() {
			_, err := normal.DisableTemporarily(&t1, nil, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "temporarily disabled period is missing the end timestamp")
		})

		Convey("temporarily disabled period must have start < end", func() {
			_, err := normal.DisableTemporarily(&t2, &t1, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the start timestamp of temporarily disabled period must be less than the end timestamp of temporarily disabled period")
		})

		Convey("account valid period must have start < end", func() {
			withAccountValidFrom_t2, err := normal.SetAccountValidFrom(&t2)
			So(err, ShouldBeNil)

			_, err = withAccountValidFrom_t2.SetAccountValidUntil(&t2)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the start timestamp of account valid period must be less than the end timestamp of the account valid period")
		})

		Convey("account_valid_from < temporarily_disabled_from", func() {
			withAccountValidFrom_t2, err := normal.SetAccountValidFrom(&t2)
			So(err, ShouldBeNil)

			_, err = withAccountValidFrom_t2.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period")
		})

		Convey("temporarily_disabled_until < account_valid_until", func() {
			withAccountValidUntil_t2, err := normal.SetAccountValidUntil(&t2)
			So(err, ShouldBeNil)

			_, err = withAccountValidUntil_t2.DisableTemporarily(&t2, &t3, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the end timestamp of temporarily disabled period must be less than the end timestamp of account valid period")
		})

		Convey("can set account_valid_from when it is temporarily disabled", func() {
			disabled, err := normal.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)

			_, err = disabled.SetAccountValidFrom(&t0)
			So(err, ShouldBeNil)
		})

		Convey("can set account_valid_until when it is temporarily disabled", func() {
			disabled, err := normal.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)

			_, err = disabled.SetAccountValidUntil(&t3)
			So(err, ShouldBeNil)
		})

		Convey("can set account valid period when it is temporarily disabled", func() {
			disabled, err := normal.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)

			_, err = disabled.SetAccountValidPeriod(&t0, &t3)
			So(err, ShouldBeNil)
		})

		Convey("can disable temporarily when it is within account valid period", func() {
			withinAccountValidPeriod, err := normal.SetAccountValidPeriod(&t0, &t3)
			So(err, ShouldBeNil)

			_, err = withinAccountValidPeriod.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)
		})

		Convey("can disable temporarily when it is before account valid period", func() {
			status, err := AccountStatus{}.WithRefTime(t0).SetAccountValidPeriod(&t1, &t4)
			So(err, ShouldBeNil)

			_, err = status.DisableTemporarily(&t2, &t3, nil)
			So(err, ShouldBeNil)
		})

		Convey("can disable temporarily when it is after account valid period", func() {
			status, err := AccountStatus{}.WithRefTime(t4).SetAccountValidPeriod(&t0, &t3)
			So(err, ShouldBeNil)

			_, err = status.DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)
		})

		Convey("past temporarily disabled period is ignored when setting account valid period", func() {
			status, err := AccountStatus{}.WithRefTime(t2).DisableTemporarily(&t0, &t1, nil)
			So(err, ShouldBeNil)
			So(status.Check(), ShouldBeNil)

			_, err = status.SetAccountValidPeriod(&t2, &t3)
			So(err, ShouldBeNil)
		})

		Convey("upcoming temporarily disabled period is NOT ignored when setting account valid period", func() {
			status, err := AccountStatus{}.WithRefTime(t0).DisableTemporarily(&t1, &t2, nil)
			So(err, ShouldBeNil)
			So(status.Check(), ShouldBeNil)

			_, err = status.SetAccountValidPeriod(&t3, &t4)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period")
		})

		Convey("ongoing temporarily disabled period is NOT ignored when setting account valid period", func() {
			status, err := AccountStatus{}.WithRefTime(t1).DisableTemporarily(&t0, &t2, nil)
			So(err, ShouldBeNil)
			So(status.Check(), ShouldBeError, "user is disabled")

			_, err = status.SetAccountValidPeriod(&t3, &t4)
			So(err, ShouldNotBeNil)
			So(err, ShouldBeError, "the start timestamp of account valid period must be less than the start timestamp of temporarily disabled period")
		})
	})
}

func TestAccountStatusPrecedence(t *testing.T) {
	Convey("AccountStatus precedence", t, func() {
		t0 := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		t1 := time.Date(2006, 1, 2, 3, 4, 5, 6+1, time.UTC)
		t2 := time.Date(2006, 1, 2, 3, 4, 5, 6+2, time.UTC)
		t3 := time.Date(2006, 1, 2, 3, 4, 5, 6+3, time.UTC)
		t4 := time.Date(2006, 1, 2, 3, 4, 5, 6+4, time.UTC)

		Convey("anonymized > account valid period", func() {
			true_ := true
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				accountValidFrom:       &t3,
				accountValidUntil:      &t4,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is anonymized")

			accountStatus.accountStatus.isAnonymized = &false_
			So(accountStatus.Check(), ShouldBeError, "user is outside valid period")
		})

		Convey("anonymised > scheduled deletion", func() {
			true_ := true
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				deleteAt:               &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is anonymized")

			accountStatus.accountStatus.isAnonymized = &false_
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by admin")

			accountStatus.accountStatus.isDeactivated = &true_
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by end-user")
		})

		Convey("anonymized > scheduled anonymization", func() {
			true_ := true
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				anonymizeAt:            &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is anonymized")

			accountStatus.accountStatus.isAnonymized = &false_
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for anonymization by admin")
		})

		Convey("anonymized > disabled indefinitely", func() {
			true_ := true
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is anonymized")

			accountStatus.accountStatus.isAnonymized = &false_
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("anonymized > disabled temporarily", func() {
			true_ := true
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:               true,
				isIndefinitelyDisabled:   &false_,
				isAnonymized:             &true_,
				temporarilyDisabledFrom:  &t1,
				temporarilyDisabledUntil: &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is anonymized")

			accountStatus.accountStatus.isAnonymized = &false_
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("account valid period > scheduled deletion", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				accountValidFrom:       &t3,
				accountValidUntil:      &t4,
				deleteAt:               &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is outside valid period")

			accountStatus.accountStatus.accountValidFrom = nil
			accountStatus.accountStatus.accountValidUntil = nil
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by admin")

			accountStatus.accountStatus.isDeactivated = &true_
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by end-user")
		})

		Convey("account valid period > scheduled anonymization", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				accountValidFrom:       &t3,
				accountValidUntil:      &t4,
				anonymizeAt:            &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is outside valid period")

			accountStatus.accountStatus.accountValidFrom = nil
			accountStatus.accountStatus.accountValidUntil = nil
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for anonymization by admin")
		})

		Convey("account valid period > disabled indefinitely", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				accountValidFrom:       &t3,
				accountValidUntil:      &t4,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user is outside valid period")

			accountStatus.accountStatus.accountValidFrom = nil
			accountStatus.accountStatus.accountValidUntil = nil
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("account valid period > disabled temporarily", func() {
			false_ := false
			accountStatus := AccountStatus{
				isDisabled:               true,
				isIndefinitelyDisabled:   &false_,
				accountValidFrom:         &t0,
				accountValidUntil:        &t3,
				temporarilyDisabledFrom:  &t1,
				temporarilyDisabledUntil: &t4,
			}.WithRefTime(t3)

			So(accountStatus.Check(), ShouldBeError, "user is outside valid period")

			accountStatus.accountStatus.accountValidFrom = nil
			accountStatus.accountStatus.accountValidUntil = nil
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("scheduled deletion > scheduled anonymization", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				deleteAt:               &t3,
				anonymizeAt:            &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by admin")

			accountStatus.accountStatus.deleteAt = nil
			So(accountStatus.Check(), ShouldBeError, "user was scheduled for anonymization by admin")
		})

		Convey("scheduled deletion > disabled indefinitely", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				deleteAt:               &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user was scheduled for deletion by admin")

			accountStatus.accountStatus.deleteAt = nil
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("scheduled deletion > disabled temporarily", func() {
			// These two statuses never overlap.
		})

		Convey("scheduled anonymization > disabled indefinitely", func() {
			true_ := true
			accountStatus := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				anonymizeAt:            &t3,
			}.WithRefTime(t2)

			So(accountStatus.Check(), ShouldBeError, "user was scheduled for anonymization by admin")

			accountStatus.accountStatus.anonymizeAt = nil
			So(accountStatus.Check(), ShouldBeError, "user is disabled")
		})

		Convey("scheduled anonymization > disabled temporarily", func() {
			// These two statues never overlap.
		})

		Convey("disabled indefinitely > disabled temporarily", func() {
			// These two statues never overlap.
		})
	})
}
