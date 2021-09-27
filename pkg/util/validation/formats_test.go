package validation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatPhone(t *testing.T) {
	f := FormatPhone{}.CheckFormat

	Convey("FormatPhone", t, func() {
		So(f(1), ShouldBeNil)
		So(f("+85298765432"), ShouldBeNil)
		So(f(""), ShouldBeError, "not in E.164 format")
		So(f("foobar"), ShouldBeError, "not in E.164 format")
	})
}

func TestFormatEmail(t *testing.T) {
	Convey("FormatEmail", t, func() {
		f := FormatEmail{}.CheckFormat

		So(f(1), ShouldBeNil)
		So(f("user@example.com"), ShouldBeNil)
		So(f(""), ShouldBeError, "mail: no address")
		So(f("John Doe <user@example.com>"), ShouldBeError, "input email must not have name")
		So(f("foobar"), ShouldBeError, "mail: missing '@' or angle-addr")
	})

	Convey("FormatEmail with name", t, func() {
		f := FormatEmail{AllowName: true}.CheckFormat

		So(f(1), ShouldBeNil)
		So(f("user@example.com"), ShouldBeNil)
		So(f("John Doe <user@example.com>"), ShouldBeNil)
		So(f(""), ShouldBeError, "mail: no address")
		So(f("foobar"), ShouldBeError, "mail: missing '@' or angle-addr")
	})
}

func TestFormatURI(t *testing.T) {
	Convey("FormatURI", t, func() {
		f := FormatURI{}.CheckFormat

		So(f(1), ShouldBeNil)
		So(f(""), ShouldBeError, "input URL must be absolute")
		So(f("/a"), ShouldBeError, "input URL must be absolute")
		So(f("#a"), ShouldBeError, "input URL must be absolute")
		So(f("?a"), ShouldBeError, "input URL must be absolute")

		So(f("http://example.com"), ShouldBeNil)
		So(f("http://example.com/"), ShouldBeNil)
		So(f("http://example.com/a"), ShouldBeNil)
		So(f("http://example.com/a/"), ShouldBeNil)

		So(f("http://example.com?"), ShouldBeNil)
		So(f("http://example.com?a"), ShouldBeNil)
		So(f("http://example.com?a=b"), ShouldBeNil)

		So(f("http://example.com#"), ShouldBeNil)
		So(f("http://example.com#a"), ShouldBeNil)
	})
}

func TestFormatHTTPOrigin(t *testing.T) {
	Convey("FormatHTTPOrigin", t, func() {
		f := FormatHTTPOrigin{}.CheckFormat

		So(f(1), ShouldBeNil)
		So(f("http://example.com"), ShouldBeNil)
		So(f("http://example.com?"), ShouldBeNil)
		So(f("http://example.com#"), ShouldBeNil)

		So(f(""), ShouldBeError, "expect input URL with scheme http / https")
		So(f("http://user:password@example.com"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f("http://example.com/"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f("http://example.com?a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f("http://example.com#a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
	})
}

func TestFormatWeChatAccountID(t *testing.T) {
	Convey("TestFormatWeChatAccountID", t, func() {
		f := FormatWeChatAccountID{}.CheckFormat

		So(f(1), ShouldBeNil)
		So(f("foobar"), ShouldBeError, "expect WeChat account id start with gh_")
		So(f("gh_foobar"), ShouldBeNil)
	})
}

func TestFormatBCP47(t *testing.T) {
	f := FormatBCP47{}.CheckFormat

	Convey("FormatBCP47", t, func() {
		So(f(1), ShouldBeNil)

		So(f(""), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f("a"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f("foobar"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")

		So(f("en"), ShouldBeNil)
		So(f("zh-TW"), ShouldBeNil)
		So(f("und"), ShouldBeNil)

		So(f("zh_TW"), ShouldBeError, "non-canonical BCP 47 tag: zh_TW != zh-TW")
	})
}
