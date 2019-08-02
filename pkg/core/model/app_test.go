package model

import (
	"net/http"
	"testing"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAccessToken(t *testing.T) {
	Convey("GetAccessToken", t, func() {
		Convey("should return value of access token header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "access-token")
			req.Header.Add(httpHeaderAuthorization, "")
			So(GetAccessToken(req), ShouldEqual, "access-token")
		})
		Convey("should return value of authorization header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "")
			req.Header.Add(httpHeaderAuthorization, "Bearer bearer-token")
			So(GetAccessToken(req), ShouldEqual, "bearer-token")
		})
		Convey("should prioritize authorization header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "access-token")
			req.Header.Add(httpHeaderAuthorization, "Bearer bearer-token")
			So(GetAccessToken(req), ShouldEqual, "bearer-token")
		})
	})
}
