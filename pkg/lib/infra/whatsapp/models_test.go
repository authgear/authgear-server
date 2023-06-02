package whatsapp_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
)

func TestLoginResponseUserExpiresTime(t *testing.T) {
	Convey("LoginResponseUserExpiresTime", t, func() {
		fixture := "2023-06-07 10:18:01+00:00"
		expectedTime, err := time.Parse(time.RFC3339, "2023-06-07T10:18:01+00:00")
		if err != nil {
			panic(err)
		}

		Convey("UnmarshalText", func() {
			var obj whatsapp.LoginResponseUserExpiresTime
			err := obj.UnmarshalText([]byte(fixture))
			So(err, ShouldBeNil)

			So(time.Time(obj), ShouldEqual, expectedTime)
		})

		Convey("MarshalText", func() {
			obj := whatsapp.LoginResponseUserExpiresTime(expectedTime)

			textbytes, err := obj.MarshalText()
			So(err, ShouldBeNil)

			So(string(textbytes), ShouldEqual, fixture)
		})

	})
}
