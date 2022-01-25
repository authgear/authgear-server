package user

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
		deleteAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
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
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletion)

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
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletion)

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
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion -> disabled")

			_, err = scheduledDeletion.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion -> normal")

			_, err = scheduledDeletion.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion -> scheduled_deletion")

			normal, err := scheduledDeletion.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)
		})
	})
}
