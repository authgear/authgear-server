package user

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type accountStatusStateTransitionTest struct {
	Reenable                       string
	Disable                        string
	ScheduleDeletionByEndUser      string
	ScheduleDeletionByAdmin        string
	UnscheduleDeletionByAdmin      string
	Anonymize                      string
	ScheduleAnonymizationByAdmin   string
	UnscheduleAnonymizationByAdmin string
}

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
		deleteAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		anonymizeAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
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

		testStateTransition := func(status AccountStatusWithRefTime, testCase accountStatusStateTransitionTest) {
			var err error

			_, err = status.Reenable()
			if testCase.Reenable == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.Reenable)
			}

			_, err = status.DisableIndefinitely(nil)
			if testCase.Disable == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.Disable)
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
				Disable:                        "",
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
				Disable:                        "invalid account status transition: disabled -> disabled",
				ScheduleDeletionByEndUser:      "invalid account status transition: disabled -> scheduled_deletion_deactivated",
				ScheduleDeletionByAdmin:        "",
				UnscheduleDeletionByAdmin:      "invalid account status transition: disabled -> normal",
				Anonymize:                      "",
				ScheduleAnonymizationByAdmin:   "",
				UnscheduleAnonymizationByAdmin: "invalid account status transition: disabled -> normal",
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
				Disable:                        "invalid account status transition: scheduled_deletion_disabled -> disabled",
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
				Disable:                        "invalid account status transition: scheduled_deletion_deactivated -> disabled",
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
				Disable:                        "invalid account status transition: anonymized -> disabled",
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
				Disable:                        "invalid account status transition: scheduled_anonymization_disabled -> disabled",
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
	})
}
