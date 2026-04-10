package smtp

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildTestMessage(t *testing.T) {
	Convey("buildTestMessage", t, func() {
		baseOpts := SendTestEmailOptions{
			To:           "user@example.com",
			SMTPHost:     "localhost",
			SMTPPort:     587,
			SMTPUsername: "user",
			SMTPPassword: "pass",
		}

		Convey("ASCII sender name", func() {
			baseOpts.SMTPSender = "Acme Corp <noreply@example.com>"
			msg, err := buildTestMessage("test-app", baseOpts)
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldContainSubstring, "<noreply@example.com>")
		})

		Convey("non-ASCII sender name", func() {
			baseOpts.SMTPSender = "Acme Corp 有限公司 <noreply@example.com>"
			msg, err := buildTestMessage("test-app", baseOpts)
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldContainSubstring, "<noreply@example.com>")
			So(strings.HasPrefix(from[0], "=?UTF-8?"), ShouldBeTrue)
		})

		Convey("invalid sender returns error", func() {
			baseOpts.SMTPSender = "not-an-email"
			_, err := buildTestMessage("test-app", baseOpts)
			So(err, ShouldNotBeNil)
		})
	})
}
