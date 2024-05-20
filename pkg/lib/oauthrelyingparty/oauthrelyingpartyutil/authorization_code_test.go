package oauthrelyingpartyutil

import (
	"testing"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCode(t *testing.T) {
	Convey("GetCode", t, func() {
		Convey("query without ?", func() {
			code, err := GetCode("code=code")
			So(err, ShouldBeNil)
			So(code, ShouldEqual, "code")

			code, err = GetCode("error=error")
			So(err, ShouldResemble, &oauthrelyingparty.ErrorResponse{
				Error_: "error",
			})
			So(code, ShouldEqual, "")
		})

		Convey("query with ?", func() {
			code, err := GetCode("?code=code")
			So(err, ShouldBeNil)
			So(code, ShouldEqual, "code")

			code, err = GetCode("?error=error")
			So(err, ShouldResemble, &oauthrelyingparty.ErrorResponse{
				Error_: "error",
			})
			So(code, ShouldEqual, "")
		})

		Convey("empty query results in empty code", func() {
			code, err := GetCode("")
			So(err, ShouldBeNil)
			So(code, ShouldEqual, "")
		})
	})
}
