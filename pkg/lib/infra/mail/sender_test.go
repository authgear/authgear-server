package mail

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/gomail.v2"
)

func TestSetFromHeader(t *testing.T) {
	Convey("SetFromHeader", t, func() {
		newMsg := func() *gomail.Message { return gomail.NewMessage() }

		Convey("no display name", func() {
			msg := newMsg()
			err := SetFromHeader(msg, "noreply@example.com")
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldEqual, "noreply@example.com")
		})

		Convey("angle-bracket addr-spec without display name", func() {
			msg := newMsg()
			err := SetFromHeader(msg, "<noreply@example.com>")
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldContainSubstring, "noreply@example.com")
		})

		Convey("ASCII-only display name", func() {
			msg := newMsg()
			err := SetFromHeader(msg, "Acme Corp <noreply@example.com>")
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldContainSubstring, "<noreply@example.com>")
		})

		Convey("display name requiring quoting (comma)", func() {
			msg := newMsg()
			err := SetFromHeader(msg, `"Doe, John" <noreply@example.com>`)
			So(err, ShouldBeNil)
			from := msg.GetHeader("From")
			So(len(from), ShouldEqual, 1)
			So(from[0], ShouldContainSubstring, "<noreply@example.com>")
		})

		// RFC 5322 display names may contain any Unicode text; non-ASCII must be
		// encoded as RFC 2047 encoded words covering only the name, not the address.
		nonASCIICases := []struct {
			name   string
			sender string
		}{
			// CJK
			{"Traditional Chinese", "範例公司 <noreply@example.com>"},
			{"Simplified Chinese", "示例公司 <noreply@example.com>"},
			{"Japanese hiragana/katakana/kanji", "テスト株式会社 <noreply@example.com>"},
			{"Korean", "테스트 회사 <noreply@example.com>"},
			// RTL scripts
			{"Arabic", "شركة اختبار <noreply@example.com>"},
			{"Hebrew", "חברת בדיקה <noreply@example.com>"},
			// Emoji (outside BMP — multi-byte UTF-8)
			{"Emoji", "Test 🚀 Company <noreply@example.com>"},
			// Mixed ASCII + non-ASCII
			{"Mixed ASCII and CJK", "Acme Corp 有限公司 <noreply@example.com>"},
		}

		for _, tc := range nonASCIICases {
			tc := tc
			Convey(tc.name, func() {
				msg := newMsg()
				err := SetFromHeader(msg, tc.sender)
				So(err, ShouldBeNil)
				from := msg.GetHeader("From")
				So(len(from), ShouldEqual, 1)
				// Email address must appear as a literal outside any encoded word.
				// The original bug encoded the entire string including "<addr>" inside
				// an RFC 2047 word, making the header unparseable.
				So(from[0], ShouldContainSubstring, "<noreply@example.com>")
				// Display name must be RFC 2047 encoded (UTF-8).
				So(strings.HasPrefix(from[0], "=?UTF-8?"), ShouldBeTrue)
			})
		}

		Convey("invalid address returns error", func() {
			msg := newMsg()
			err := SetFromHeader(msg, "not-an-email-address")
			So(err, ShouldNotBeNil)
		})
	})
}
