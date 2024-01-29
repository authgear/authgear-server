package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIndentRule(t *testing.T) {
	Convey("Given an HTML string with tabs for indentation", t, func() {
		html := `<html>
	<head>
		<title>Test</title>
	</head>
	<body>
		<p>Test</p>
	</body>
</html>`
		violations := IndentationRule{}.Check(html, "test.html")
		Convey("Then I expect to see a lint violation", func() {
			So(len(violations), ShouldEqual, 6)
			So(violations[0].Path, ShouldEqual, "test.html")
			So(violations[0].Line, ShouldEqual, 2)
			So(violations[0].Column, ShouldEqual, 1)
			So(violations[0].Message, ShouldEqual, "Indentation is a tab instead of 2 spaces (indent)")
		})
	})

	Convey("Given an HTML string with spaces for indentation", t, func() {
		html := `<html>
  <head>
    <title>Test</title>
  </head>
  <body>
    <p>Test</p>
  </body>
</html>`
		violations := IndentationRule{}.Check(html, "test.html")
		Convey("Then I expect to see no lint violations", func() {
			So(len(violations), ShouldEqual, 0)
		})
	})

	Convey("Given an HTML string with tabs in between spaces for indentation", t, func() {
		html := `<html>
  <head>
  	<title>Test</title><!-- 2 spaces + 1 tab -->
  </head>
	<body><!-- 1 tab -->
    <p>Test</p>
  </body>
</html>`

		violations := IndentationRule{}.Check(html, "test.html")
		Convey("Then I expect to see a lint violation", func() {
			So(len(violations), ShouldEqual, 2)
			So(violations[0].Path, ShouldEqual, "test.html")
			So(violations[0].Line, ShouldEqual, 3)
			So(violations[0].Column, ShouldEqual, 3)
			So(violations[0].Message, ShouldEqual, "Indentation is a tab instead of 2 spaces (indent)")
		})
	})

	Convey("Given an HTML string with tab but not for indentation", t, func() {
		html := `<html>
  <head>
    <title>Test</title>
  </head>	<!-- 1 tab at the end -->
  <body>
    <p class="	tab">Test</p><!-- 1 tab in attribute -->
  </body>
</html>`
		violations := IndentationRule{}.Check(html, "test.html")
		Convey("Then I expect to see no lint violations", func() {
			So(len(violations), ShouldEqual, 0)
		})
	})
}
