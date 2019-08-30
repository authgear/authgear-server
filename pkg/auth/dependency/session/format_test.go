package session

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseUserAgent(t *testing.T) {
	Convey("parseUserAgent", t, func() {
		Convey("should parse browser UA correctly", func() {
			ua := parseUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
			So(ua, ShouldResemble, model.SessionUserAgent{
				Raw:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36",
				Name:        "Chrome",
				Version:     "75.0.3770",
				OS:          "Mac OS X",
				OSVersion:   "10.14.5",
				DeviceModel: "",
			})
		})
		Convey("should parse Skygear SDK UA correctly", func() {
			ua := parseUserAgent("io.skygear.test/1.0.1 (Skygear; iPhone11,8; iOS 12.0) SKYKit/2.0.1")
			So(ua, ShouldResemble, model.SessionUserAgent{
				Raw:         "io.skygear.test/1.0.1 (Skygear; iPhone11,8; iOS 12.0) SKYKit/2.0.1",
				Name:        "io.skygear.test",
				Version:     "1.0.1",
				OS:          "iOS",
				OSVersion:   "12.0",
				DeviceModel: "Apple iPhone11,8",
			})

			ua = parseUserAgent("io.skygear.test/1.3.0 (Skygear; Samsung GT-S5830L; Android 9.0) io.skygear.skygear/2.2.0")
			So(ua, ShouldResemble, model.SessionUserAgent{
				Raw:         "io.skygear.test/1.3.0 (Skygear; Samsung GT-S5830L; Android 9.0) io.skygear.skygear/2.2.0",
				Name:        "io.skygear.test",
				Version:     "1.3.0",
				OS:          "Android",
				OSVersion:   "9.0",
				DeviceModel: "Samsung GT-S5830L",
			})
		})
	})
}

func TestResolveIP(t *testing.T) {
	Convey("resolveIP", t, func() {
		Convey("should resolve X-Real-IP", func() {
			ip := resolveIP(auth.SessionAccessEventConnInfo{
				XRealIP: "169.254.198.67",
			})
			So(ip, ShouldEqual, "169.254.198.67")
		})
		Convey("should resolve X-Forwarded-For", func() {
			ip := resolveIP(auth.SessionAccessEventConnInfo{
				XForwardedFor: "[::1]:20595, 169.254.198.67",
			})
			So(ip, ShouldEqual, "::1")
		})
		Convey("should resolve Forwarded", func() {
			ip := resolveIP(auth.SessionAccessEventConnInfo{
				Forwarded: "for=127.0.0.1:313;by=169.254.198.67, for=169.254.198.67",
			})
			So(ip, ShouldEqual, "127.0.0.1")
		})
		Convey("should resolve RemoteAddr", func() {
			ip := resolveIP(auth.SessionAccessEventConnInfo{
				RemoteAddr: "1.1.1.1:7236",
			})
			So(ip, ShouldEqual, "1.1.1.1")
		})
		Convey("should resolve with priority", func() {
			ip := resolveIP(auth.SessionAccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			})
			So(ip, ShouldEqual, "a")

			ip = resolveIP(auth.SessionAccessEventConnInfo{
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			})
			So(ip, ShouldEqual, "b")

			ip = resolveIP(auth.SessionAccessEventConnInfo{
				Forwarded:  "for=c",
				RemoteAddr: "d",
			})
			So(ip, ShouldEqual, "c")

			ip = resolveIP(auth.SessionAccessEventConnInfo{
				RemoteAddr: "d",
			})
			So(ip, ShouldEqual, "d")
		})
	})
}
