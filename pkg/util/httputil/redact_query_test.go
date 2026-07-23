package httputil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestRedactedRawQuery(t *testing.T) {
	Convey("RedactedRawQuery", t, func() {
		Convey("should return the raw query unchanged when no sensitive params are present", func() {
			So(httputil.RedactedRawQuery("client_id=test&state=abc"), ShouldEqual, "client_id=test&state=abc")
		})

		Convey("should return empty string for empty raw query", func() {
			So(httputil.RedactedRawQuery(""), ShouldEqual, "")
		})

		Convey("should redact id_token_hint and login_hint", func() {
			So(
				httputil.RedactedRawQuery("id_token_hint=eyJhbGci.xxx.yyy&login_hint=https%3A%2F%2Fauthgear.com%2Flogin_hint%3Femail%3Dtest%40example.com"),
				ShouldEqual,
				"id_token_hint=redacted&login_hint=redacted",
			)
		})

		Convey("should redact code and token while preserving other params", func() {
			So(httputil.RedactedRawQuery("client_id=test&code=123456"), ShouldEqual, "client_id=test&code=redacted")
			So(httputil.RedactedRawQuery("client_id=test&token=abcdef"), ShouldEqual, "client_id=test&token=redacted")
		})
	})
}
