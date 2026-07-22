package httputil_test

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestRedactedURLString(t *testing.T) {
	Convey("RedactedURLString", t, func() {
		Convey("should not modify URL without sensitive query params", func() {
			u, _ := url.Parse("/oauth2/authorize?client_id=test&state=abc")
			So(httputil.RedactedURLString(u), ShouldEqual, "/oauth2/authorize?client_id=test&state=abc")
		})

		Convey("should redact id_token_hint and login_hint", func() {
			u, _ := url.Parse("/oauth2/end_session?id_token_hint=eyJhbGci.xxx.yyy&login_hint=https%3A%2F%2Fauthgear.com%2Flogin_hint%3Femail%3Dtest%40example.com")
			So(httputil.RedactedURLString(u), ShouldEqual, "/oauth2/end_session?id_token_hint=redacted&login_hint=redacted")
		})

		Convey("should redact code and token", func() {
			u, _ := url.Parse("/verify_login_link?code=123456")
			So(httputil.RedactedURLString(u), ShouldEqual, "/verify_login_link?code=redacted")

			u2, _ := url.Parse("/tester?token=abcdef")
			So(httputil.RedactedURLString(u2), ShouldEqual, "/tester?token=redacted")
		})

		Convey("should preserve other params while redacting sensitive ones", func() {
			u, _ := url.Parse("/oauth2/end_session?client_id=test&id_token_hint=eyJ.xxx.yyy&state=abc")
			So(httputil.RedactedURLString(u), ShouldEqual, "/oauth2/end_session?client_id=test&id_token_hint=redacted&state=abc")
		})
	})
}

func TestRedactedRawQuery(t *testing.T) {
	Convey("RedactedRawQuery", t, func() {
		Convey("should return the raw query unchanged when no sensitive params are present", func() {
			So(httputil.RedactedRawQuery("client_id=test&state=abc"), ShouldEqual, "client_id=test&state=abc")
		})

		Convey("should return empty string for empty raw query", func() {
			So(httputil.RedactedRawQuery(""), ShouldEqual, "")
		})

		Convey("should redact sensitive params", func() {
			So(httputil.RedactedRawQuery("client_id=test&id_token_hint=eyJ.xxx.yyy"), ShouldEqual, "client_id=test&id_token_hint=redacted")
		})
	})
}
