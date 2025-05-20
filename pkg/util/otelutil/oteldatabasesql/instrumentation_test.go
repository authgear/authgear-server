package oteldatabasesql

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS(t *testing.T) {
	Convey("Test get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS function", t, func() {
		Convey("When env var is not set", func() {
			os.Unsetenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(0))
		})

		Convey("When env var is empty", func() {
			os.Setenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS", "")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(0))
		})

		Convey("When env var is not a valid integer", func() {
			os.Setenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS", "not-an-integer")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(0))
		})

		Convey("When env var is zero", func() {
			os.Setenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS", "0")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(0))
		})

		Convey("When env var is negative", func() {
			os.Setenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS", "-100")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(0))
		})

		Convey("When env var is positive", func() {
			os.Setenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS", "1000")

			timeout := get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
			So(timeout, ShouldEqual, time.Duration(1000)*time.Millisecond)
		})
	})
}
