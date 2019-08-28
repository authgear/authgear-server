package model

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAccessToken(t *testing.T) {
	Convey("GetAccessToken", t, func() {

		Convey("should return value of access token header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "access-token")
			req.Header.Add(httpHeaderAuthorization, "")
			token, transport, err := GetAccessToken(req)
			So(err, ShouldBeNil)
			So(transport, ShouldEqual, config.SessionTransportTypeHeader)
			So(token, ShouldEqual, "access-token")
		})
		Convey("should return value of authorization header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "")
			req.Header.Add(httpHeaderAuthorization, "Bearer bearer-token")
			token, transport, err := GetAccessToken(req)
			So(err, ShouldBeNil)
			So(transport, ShouldEqual, config.SessionTransportTypeHeader)
			So(token, ShouldEqual, "bearer-token")
		})
		Convey("should prioritize authorization header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(corehttp.HeaderAccessToken, "access-token")
			req.Header.Add(httpHeaderAuthorization, "Bearer bearer-token")
			token, transport, err := GetAccessToken(req)
			So(err, ShouldBeNil)
			So(transport, ShouldEqual, config.SessionTransportTypeHeader)
			So(token, ShouldEqual, "bearer-token")
		})

		Convey("should return value of cookie", func() {
			req, _ := http.NewRequest("", "", nil)
			req.AddCookie(&http.Cookie{Name: corehttp.CookieNameSession, Value: "cookie-token"})
			token, transport, err := GetAccessToken(req)
			So(err, ShouldBeNil)
			So(transport, ShouldEqual, config.SessionTransportTypeCookie)
			So(token, ShouldEqual, "cookie-token")
		})
		Convey("should return error if value present in both cookie and header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(httpHeaderAuthorization, "Bearer header-token")
			req.AddCookie(&http.Cookie{Name: corehttp.CookieNameSession, Value: "cookie-token"})
			_, _, err := GetAccessToken(req)
			So(err, ShouldBeError, "tokens detected in different transports")
		})
	})
}
