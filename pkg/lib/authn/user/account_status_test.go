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
	ScheduleDeletionByEndUser      string
	ScheduleDeletionByAdmin        string
	UnscheduleDeletionByAdmin      string
	Anonymize                      string
	ScheduleAnonymizationByAdmin   string
	UnscheduleAnonymizationByAdmin string
}

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
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

			_, err = status.DisableTemporarily(temporarilyDisabledFrom, temporarilyDisabledUntil, nil)
			if testCase.DisableTemporarily_Now == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.DisableTemporarily_Now)
			}

			_, err = status.DisableTemporarily(temporarilyDisabledFromInFuture, temporarilyDisabledUntilInFuture, nil)
			if testCase.DisableTemporarily_Future == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.DisableTemporarily_Future)
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
				DisableTemporarily_Now:         "invalid account status transition: disabled_temporarily -> disabled_temporarily",
				DisableTemporarily_Future:      "invalid account status transition: disabled_temporarily -> normal",
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
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
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
				DisableTemporarily_Now:         "",
				DisableTemporarily_Future:      "",
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
				ScheduleDeletionByEndUser:      "invalid account status transition: scheduled_anonymization_disabled -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: scheduled_anonymization_disabled -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "invalid account status transition: scheduled_anonymization_disabled -> scheduled_anonymization_disabled",
				UnscheduleAnonymizationByAdmin: "",
			})
		})

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

		Convey("isDisabled is true if now is within temporarily disabled period", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.DisableTemporarily(temporarilyDisabledFrom, temporarilyDisabledUntil, nil)
			So(err, ShouldBeNil)
			So(state1.IsDisabled(), ShouldEqual, true)
			So(state1.accountStatus.isDisabled, ShouldEqual, true)
			So(*state1.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isDeactivated, ShouldEqual, false)
		})

		Convey("isDisabled is false if now is NOT within temporarily disabled period", func() {
			normal := AccountStatus{}.WithRefTime(now)

			state1, err := normal.DisableTemporarily(temporarilyDisabledFromInFuture, temporarilyDisabledUntilInFuture, nil)
			So(err, ShouldBeNil)
			So(state1.IsDisabled(), ShouldEqual, false)
			So(state1.accountStatus.isDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isIndefinitelyDisabled, ShouldEqual, false)
			So(*state1.accountStatus.isDeactivated, ShouldEqual, false)
		})

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
	})
}
