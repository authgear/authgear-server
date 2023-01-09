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
		Convey("normal", func() {
			var normal AccountStatus
			var err error

			_, err = normal.Reenable()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")

			disabled, err := normal.Disable(nil)
			So(err, ShouldBeNil)
			So(disabled.Type(), ShouldEqual, AccountStatusTypeDisabled)

			scheduledDeletion, err := normal.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = normal.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")
		})

		Convey("disable", func() {
			disabled := AccountStatus{
				IsDisabled: true,
			}
			var err error

			_, err = disabled.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: disabled -> disabled")

			normal, err := disabled.Reenable()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)

			scheduledDeletion, err := disabled.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = disabled.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: disabled -> normal")
		})

		Convey("scheduled deletion by admin", func() {
			scheduledDeletion := AccountStatus{
				IsDisabled: true,
				DeleteAt:   &deleteAt,
			}
			var err error

			_, err = scheduledDeletion.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> disabled")

			_, err = scheduledDeletion.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> normal")

			_, err = scheduledDeletion.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> scheduled_deletion_disabled")

			normal, err := scheduledDeletion.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)
		})

		Convey("anonymize", func() {
			anonymized := AccountStatus{
				IsDisabled:   true,
				IsAnonymized: true,
			}
			var err error

			_, err = anonymized.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: anonymized -> disabled")

			_, err = anonymized.Reenable()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> normal")

			_, err = anonymized.Anonymize()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> anonymized")
		})

		Convey("scheduled anonymization by admin", func() {
			scheduledAnonymization := AccountStatus{
				IsDisabled:  true,
				AnonymizeAt: &anonymizeAt,
			}
			var err error

			_, err = scheduledAnonymization.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> disabled")

			_, err = scheduledAnonymization.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> normal")

			_, err = scheduledAnonymization.ScheduleAnonymizationByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> scheduled_anonymization_disabled")

			normal, err := scheduledAnonymization.UnscheduleAnonymizationByAdmin()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)
		})
	})
}
