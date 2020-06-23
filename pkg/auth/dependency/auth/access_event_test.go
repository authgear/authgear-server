package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAccessEvent(t *testing.T) {
	Convey("NewAccessEvent", t, func() {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		Convey("should record current timestamp", func() {
			req, _ := http.NewRequest("POST", "", nil)

			event := NewAccessEvent(now, req, true)
			So(event.Timestamp, ShouldResemble, now)
		})
		Convey("should populate connection info", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("X-Forwarded-For", "13.225.103.28, 216.58.197.110")
			req.Header.Set("X-Real-IP", "216.58.197.110")
			req.Header.Set("Forwarded", "for=216.58.197.110;proto=http;by=192.168.1.11")

			event := NewAccessEvent(now, req, true)
			So(event.RemoteIP, ShouldResemble, "216.58.197.110")
			event = NewAccessEvent(now, req, false)
			So(event.RemoteIP, ShouldResemble, "192.168.1.11")
		})
		Convey("should populate user agent", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("User-Agent", "SDK")

			event := NewAccessEvent(now, req, true)
			So(event.UserAgent, ShouldEqual, "SDK")
		})
		Convey("should populate extra info", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.Header.Set("X-Authgear-Extra-Info", "eyAiZGV2aWNlX25hbWUiOiAiRGV2aWNlIiB9")

			event := NewAccessEvent(now, req, true)
			So(event.Extra, ShouldResemble, AccessEventExtraInfo{
				"device_name": "Device",
			})
		})
		Convey("should not populate extra info if too large", func() {
			extra := "{ "
			for i := 0; i < 1000; i++ {
				if i != 0 {
					extra += ", "
				}
				extra += fmt.Sprintf(`"info_%d": %d`, i, i)
			}
			extra += " }"
			extra = base64.StdEncoding.EncodeToString([]byte(extra))

			req, _ := http.NewRequest("POST", "", nil)
			req.Header.Set("X-Skygear-Extra-Info", extra)

			event := NewAccessEvent(now, req, true)
			So(event.Extra, ShouldResemble, AccessEventExtraInfo{})
		})
	})
}

func TestAccessEventConnInfoIP(t *testing.T) {
	Convey("AccessEventConnInfo.IP", t, func() {
		Convey("should resolve X-Real-IP", func() {
			ip := AccessEventConnInfo{
				XRealIP: "169.254.198.67",
			}.IP(true)
			So(ip, ShouldEqual, "169.254.198.67")
		})
		Convey("should resolve X-Forwarded-For", func() {
			ip := AccessEventConnInfo{
				XForwardedFor: "[::1]:20595, 169.254.198.67",
			}.IP(true)
			So(ip, ShouldEqual, "::1")
		})
		Convey("should resolve Forwarded", func() {
			ip := AccessEventConnInfo{
				Forwarded: "for=127.0.0.1:313;by=169.254.198.67, for=169.254.198.67",
			}.IP(true)
			So(ip, ShouldEqual, "127.0.0.1")
		})
		Convey("should resolve RemoteAddr", func() {
			ip := AccessEventConnInfo{
				RemoteAddr: "1.1.1.1:7236",
			}.IP(true)
			So(ip, ShouldEqual, "1.1.1.1")
		})
		Convey("should resolve with priority", func() {
			ip := AccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			}.IP(true)
			So(ip, ShouldEqual, "c")

			ip = AccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				RemoteAddr:    "d",
			}.IP(true)
			So(ip, ShouldEqual, "b")

			ip = AccessEventConnInfo{
				XRealIP:    "a",
				RemoteAddr: "d",
			}.IP(true)
			So(ip, ShouldEqual, "a")

			ip = AccessEventConnInfo{
				RemoteAddr: "d",
			}.IP(true)
			So(ip, ShouldEqual, "d")
		})
		Convey("should ignore headers when not trusting proxy", func() {
			ip := AccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			}.IP(false)
			So(ip, ShouldEqual, "d")

			ip = AccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				RemoteAddr:    "d",
			}.IP(false)
			So(ip, ShouldEqual, "d")

			ip = AccessEventConnInfo{
				XRealIP:    "a",
				RemoteAddr: "d",
			}.IP(false)
			So(ip, ShouldEqual, "d")

			ip = AccessEventConnInfo{
				RemoteAddr: "d",
			}.IP(false)
			So(ip, ShouldEqual, "d")
		})
	})
}
