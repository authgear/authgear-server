package user

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
		deleteAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		anonymizeAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		now := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		Convey("normal", func() {
			normal := AccountStatus{}.WithRefTime(now)
			var err error

			_, err = normal.Reenable()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")

			disabled, err := normal.Disable(nil)
			So(err, ShouldBeNil)
			So(disabled.Variant().Type(), ShouldEqual, AccountStatusTypeDisabled)

			scheduledDeletion, err := normal.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Variant().Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = normal.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")
		})

		Convey("disable", func() {
			true_ := true
			disabled := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
			}.WithRefTime(now)
			var err error

			_, err = disabled.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: disabled -> disabled")

			normal, err := disabled.Reenable()
			So(err, ShouldBeNil)
			So(normal.Variant().Type(), ShouldEqual, AccountStatusTypeNormal)

			scheduledDeletion, err := disabled.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Variant().Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = disabled.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: disabled -> normal")
		})

		Convey("scheduled deletion by admin", func() {
			true_ := true
			scheduledDeletion := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				deleteAt:               &deleteAt,
			}.WithRefTime(now)
			var err error

			_, err = scheduledDeletion.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> disabled")

			_, err = scheduledDeletion.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> normal")

			_, err = scheduledDeletion.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> scheduled_deletion_disabled")

			normal, err := scheduledDeletion.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(normal.Variant().Type(), ShouldEqual, AccountStatusTypeNormal)
		})

		Convey("anonymize", func() {
			true_ := true
			anonymized := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				isAnonymized:           &true_,
				anonymizeAt:            &anonymizeAt,
			}.WithRefTime(now)
			var err error

			_, err = anonymized.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: anonymized -> disabled")

			_, err = anonymized.Reenable()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> normal")

			_, err = anonymized.Anonymize()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> anonymized")
		})

		Convey("scheduled anonymization by admin", func() {
			true_ := true
			scheduledAnonymization := AccountStatus{
				isDisabled:             true,
				isIndefinitelyDisabled: &true_,
				anonymizeAt:            &anonymizeAt,
			}.WithRefTime(now)
			var err error

			_, err = scheduledAnonymization.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> disabled")

			_, err = scheduledAnonymization.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> normal")

			_, err = scheduledAnonymization.ScheduleAnonymizationByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> scheduled_anonymization_disabled")

			_, err = scheduledAnonymization.ScheduleDeletionByEndUser(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> scheduled_deletion_deactivated")

			anonymized, err := scheduledAnonymization.Anonymize()
			So(err, ShouldBeNil)
			So(anonymized.Variant().Type(), ShouldEqual, AccountStatusTypeAnonymized)

			unscheduleAnonymization, err := scheduledAnonymization.UnscheduleAnonymizationByAdmin()
			So(err, ShouldBeNil)
			So(unscheduleAnonymization.Variant().Type(), ShouldEqual, AccountStatusTypeNormal)

			scheduleDeletion, err := scheduledAnonymization.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduleDeletion.Variant().Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)
		})
	})
}
