package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateCookie(t *testing.T) {
	Convey("updateCookie", t, func() {
		Convey("should set new cookie", func() {
			rw := httptest.NewRecorder()
			updateCookie(rw, &http.Cookie{Name: "Test", Value: "Value1"})

			cookies := rw.Result().Cookies()
			So(cookies, ShouldResemble, []*http.Cookie{
				&http.Cookie{Name: "Test", Value: "Value1", Raw: `Test=Value1`},
			})
		})
		Convey("should update existing cookie", func() {
			rw := httptest.NewRecorder()
			updateCookie(rw, &http.Cookie{Name: "Test", Value: "Value1", Path: "/"})
			updateCookie(rw, &http.Cookie{Name: "Test", Value: "Value2", Path: "/"})

			cookies := rw.Result().Cookies()
			So(cookies, ShouldResemble, []*http.Cookie{
				&http.Cookie{Name: "Test", Value: "Value2", Path: "/", Raw: `Test=Value2; Path=/`},
			})
		})
	})
}
