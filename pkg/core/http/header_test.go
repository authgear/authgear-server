package http

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveSkygearHeader(t *testing.T) {
	Convey("RemoveSkygearHeader", t, func() {
		So(RemoveSkygearHeader(http.Header{
			"x-skygear-a": {},
			"X-Skygear-B": {},
			"X-SKYGEAR-C": {},
			"x-skygeaR-D": {},
			"host":        {},
			"Accept":      {},
		}), ShouldResemble, http.Header{
			"host":   {},
			"Accept": {},
		})
	})
}

func TestGetSessionIdentifier(t *testing.T) {
	Convey("GetSessionIdentifier", t, func() {

		Convey("should return value of authorization header", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(httpHeaderAuthorization, "Bearer bearer-token")
			token := GetSessionIdentifier(req)
			So(token, ShouldEqual, "bearer-token")
		})

		Convey("should return value of cookie", func() {
			req, _ := http.NewRequest("", "", nil)
			req.AddCookie(&http.Cookie{Name: CookieNameSession, Value: "cookie-token"})
			token := GetSessionIdentifier(req)
			So(token, ShouldEqual, "cookie-token")
		})
		Convey("should return value of cookie if both are present", func() {
			req, _ := http.NewRequest("", "", nil)
			req.Header.Add(httpHeaderAuthorization, "Bearer header-token")
			req.AddCookie(&http.Cookie{Name: CookieNameSession, Value: "cookie-token"})
			token := GetSessionIdentifier(req)
			So(token, ShouldEqual, "cookie-token")
		})
	})
}
