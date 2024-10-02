package theme

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTheme(t *testing.T) {
	Convey("CheckDeclarationInSelector", t, func() {
		test := func(cssString string, selector string, declarationProperty string, expected bool) {
			actual, _ := CheckDeclarationInSelector(cssString, selector, declarationProperty)
			So(actual, ShouldEqual, expected)
		}

		Convey("empty string should be false", func() {
			test("", ":root", "--brand-logo__height", false)
			test("", "randomSelector", "randomProperty", false)
		})

		Convey("none matched property should be false", func() {
			test(":root{}", ":root", "--brand-logo__height", false)
			test(`:root{
			--layout__bg-color: #F0FFFF;
}`, ":root", "--brand-logo__height", false)
		})

		Convey("matched property should be true", func() {
			test(`:root{
			--brand-logo__height: 40px;
}`, ":root", "--brand-logo__height", true)
		})
	})
}
