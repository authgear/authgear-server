package validation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestURL(t *testing.T) {
	Convey("URL", t, func() {
		Convey("AllowAbsoluteURI = false", func(_ C) {
			cases := []struct {
				input    string
				expected bool
			}{
				{"", false},
				{"a", false},
				{"我", false},
				{"/?a", false},
				{"/#a", false},
				{"http://example.com", false},
				{"../a", false},
				{"/a/..", false},

				{"/", true},
				{"/a", true},
				{"/a/b", true},
				{"/%20", true},
				{"/Hong+Kong", true},
			}
			for _, c := range cases {
				actual := URL{}.IsFormat(c.input)
				So(actual, ShouldEqual, c.expected)
			}
		})
		Convey("AllowAbsoluteURI = true", func(_ C) {
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

				{"/", true},
				{"/a", true},
				{"/a/b", true},
				{"/%20", true},
				{"/Hong+Kong", true},
				{"http://example.com", true},
				{"http://example.com/", true},
				{"http://example.com/a", true},
				{"http://example.com/a/b", true},
			}
			for _, c := range cases {
				actual := URL{AllowAbsoluteURI: true}.IsFormat(c.input)
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
