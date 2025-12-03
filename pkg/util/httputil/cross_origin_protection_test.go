package httputil

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCrossOriginProtection(t *testing.T) {
	convey.Convey("CrossOriginProtection", t, func() {
		cop := &CrossOriginProtection{}

		convey.Convey("Check", func() {
			convey.Convey("should return nil for safe methods", func() {
				req, _ := http.NewRequest("GET", "https://example.com", nil)
				err := cop.Check(req)
				convey.So(err, convey.ShouldBeNil)
			})

			convey.Convey("should return nil for unsafe methods with same-origin", func() {
				req, _ := http.NewRequest("POST", "https://example.com", nil)
				req.Header.Set("Sec-Fetch-Site", "same-origin")
				err := cop.Check(req)
				convey.So(err, convey.ShouldBeNil)
			})

			convey.Convey("should return error for unsafe methods with cross-site and untrusted origin", func() {
				req, _ := http.NewRequest("POST", "https://example.com", nil)
				req.Header.Set("Sec-Fetch-Site", "cross-site")
				req.Header.Set("Origin", "https://untrusted.com")
				err := cop.Check(req)
				convey.So(err, convey.ShouldNotBeNil)
			})

			convey.Convey("should return nil for unsafe methods without Sec-Fetch-Site but with same-origin Origin", func() {
				req, _ := http.NewRequest("POST", "https://example.com", nil)
				req.Header.Set("Origin", "https://example.com")
				err := cop.Check(req)
				convey.So(err, convey.ShouldBeNil)
			})

			convey.Convey("should return error for unsafe methods without Sec-Fetch-Site but with cross-origin Origin", func() {
				req, _ := http.NewRequest("POST", "https://example.com", nil)
				req.Header.Set("Origin", "https://untrusted.com")
				err := cop.Check(req)
				convey.So(err, convey.ShouldNotBeNil)
			})

			convey.Convey("should return nil for unsafe methods without Sec-Fetch-Site nor Origin", func() {
				req, _ := http.NewRequest("POST", "https://example.com", nil)
				err := cop.Check(req)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})
}
