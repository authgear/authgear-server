package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEOLAtEOFRule(t *testing.T) {
	Convey("Given a empty file", t, func() {
		html := ``
		violations := EOLAtEOFRule{}.Check(html, "test.html")
		Convey("Then I expect to see a lint violation", func() {
			So(len(violations), ShouldEqual, 0)
		})
	})

	Convey("Given a file with a single line", t, func() {
		html := `<!DOCTYPE html><html lang="en"><head><title>Test</title></head><body><p>Test</p></body></html>`
		violations := EOLAtEOFRule{}.Check(html, "test.html")
		Convey("Then I expect to see a lint violation", func() {
			So(len(violations), ShouldEqual, 1)
			So(violations[0].Path, ShouldEqual, "test.html")
			So(violations[0].Line, ShouldEqual, 1)
			So(violations[0].Column, ShouldEqual, 95)
			So(violations[0].Message, ShouldEqual, "File does not end with a newline (final-newline)")
		})
	})

	Convey("Given a file with a single line and a newline", t, func() {
		html := `<!DOCTYPE html><html lang="en"><head><title>Test</title></head><body><p>Test</p></body></html>
`
		violations := EOLAtEOFRule{}.Check(html, "test.html")
		Convey("Then I expect to see no violation", func() {
			So(len(violations), ShouldEqual, 0)
		})
	})

	Convey("Given a file with a single character and a newline", t, func() {
		html := `A
`
		violations := EOLAtEOFRule{}.Check(html, "test.html")
		Convey("Then I expect to see no violation", func() {
			So(len(violations), ShouldEqual, 0)
		})
	})

	Convey("Given a file with multiple lines", t, func() {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
<title>Test</title>
</head>
<body>
<p>Test</p>
</body>
</html>`
		violations := EOLAtEOFRule{}.Check(html, "test.html")
		Convey("Then I expect to see a lint violation", func() {
			So(len(violations), ShouldEqual, 1)
			So(violations[0].Path, ShouldEqual, "test.html")
			So(violations[0].Line, ShouldEqual, 9)
			So(violations[0].Column, ShouldEqual, 8)
			So(violations[0].Message, ShouldEqual, "File does not end with a newline (final-newline)")
		})
	})
}
