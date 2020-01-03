package validation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestURL(t *testing.T) {
	Convey("URL", t, func() {
		Convey("URLVariant == URLVariantFullOnly", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"a", false},
				{"我", false},
				{"/?a", false},
				{"/#a", false},
				{"../a", false},
				{"/a/..", false},
				{"/", false},
				{"/a", false},
				{"/a/b", false},
				{"/%20", false},
				{"/Hong+Kong", false},

				{"http://example.com", true},
				{"http://example.com/", true},
				{"http://example.com/a", true},
			}
			for _, c := range cases {
				actual := URL{URLVariant: URLVariantFullOnly}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("URLVariant == URLVariantPathOnly", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"a", false},
				{"我", false},
				{"../a", false},
				{"/?a", false},
				{"/#a", false},
				{"/a/..", false},
				{"http://example.com/?a", false},
				{"http://example.com/#a", false},
				{"http://example.com/a/..", false},
				{"http://example.com", false},
				{"http://example.com/", false},
				{"http://example.com/a", false},
				{"http://example.com/a/b", false},

				{"/", true},
				{"/a", true},
				{"/a/b", true},
				{"/%20", true},
				{"/Hong+Kong", true},
			}
			for _, c := range cases {
				actual := URL{URLVariant: URLVariantPathOnly}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("URLVariant == URLVariantFullOrPath", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"a", false},
				{"我", false},
				{"../a", false},
				{"/?a", false},
				{"/#a", false},
				{"/a/..", false},
				{"http://example.com/?a", false},
				{"http://example.com/#a", false},
				{"http://example.com/a/..", false},

				{"http://example.com", true},
				{"http://example.com/", true},
				{"http://example.com/a", true},
				{"http://example.com/a/b", true},
				{"/", true},
				{"/a", true},
				{"/a/b", true},
				{"/%20", true},
				{"/Hong+Kong", true},
			}
			for _, c := range cases {
				actual := URL{URLVariant: URLVariantFullOrPath}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
	})
}

func TestFilePath(t *testing.T) {
	Convey("FilePath", t, func() {
		Convey("Relative = true", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"/", false},

				{"a", true},
				{"./a", true},
				{"../a", true},
			}
			for _, c := range cases {
				actual := FilePath{Relative: true}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("Relative = false", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"a", false},
				{"./a", false},
				{"../a", false},

				{"/", true},
			}
			for _, c := range cases {
				actual := FilePath{Relative: false}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("File = true", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"/", false},
				{"/a/", false},

				{"/a", true},
			}
			for _, c := range cases {
				actual := FilePath{File: true}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("RelativeDirectoryPath", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"/", false},
				{"/a", false},
				{"/a/", false},

				{"a", true},
				{"a/", true},
				{"../a", true},
				{"../a/", true},
				{"./a", true},
				{"./a/", true},
			}
			for _, c := range cases {
				actual := FilePath{Relative: true, File: false}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("RelativeFilePath", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"/", false},
				{"/a", false},
				{"/a/", false},

				{"a", true},
				{"a/", false},
				{"../a", true},
				{"../a/", false},
				{"./a", true},
				{"./a/", false},
			}
			for _, c := range cases {
				actual := FilePath{Relative: true, File: true}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
	})
}

func TestE164Phone(t *testing.T) {
	Convey("E164Phone", t, func() {
		f := E164Phone{}.IsFormat
		So(f(nil), ShouldBeTrue)
		So(f(1), ShouldBeTrue)
		So(f(""), ShouldBeTrue)
		So(f("+85222334455"), ShouldBeTrue)
		So(f(" +85222334455"), ShouldBeFalse)
		So(f("nonsense"), ShouldBeFalse)
	})
}

func TestEmail(t *testing.T) {
	Convey("Email", t, func() {
		f := Email{}.IsFormat
		So(f(nil), ShouldBeTrue)
		So(f(1), ShouldBeTrue)
		So(f(""), ShouldBeTrue)
		So(f("user@example.com"), ShouldBeTrue)
		So(f("User <user@example.com>"), ShouldBeFalse)
		So(f(" user@example.com "), ShouldBeFalse)
		So(f("nonsense"), ShouldBeFalse)
	})
}
