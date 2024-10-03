package theme

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tdewolff/parse/v2"
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

	Convey("AddDeclarationInSelectorIfNotPresentAlready", t, func() {
		test := func(cssString string, selector string, declaration Declaration, expectedCSS string, expectedAdded bool, expectedErr error) {
			newCSS, added, err := AddDeclarationInSelectorIfNotPresentAlready(cssString, selector, declaration)
			So(newCSS, ShouldEqual, expectedCSS)
			So(added, ShouldEqual, expectedAdded)
			So(err, ShouldResemble, expectedErr)
		}

		var defaultBrandLogoHeight = Declaration{Property: "--brand-logo__height", Value: "40px"}

		Convey("bad css input should return empty string and error", func() {
			test("abcdefg", ":root", Declaration{}, "", false, &parse.Error{Message: "unexpected ending in qualified rule", Line: 1, Column: 8, Context: "    1: abcdefg\n              ^"})
			test("iambad", ":root", Declaration{}, "", false, &parse.Error{Message: "unexpected ending in qualified rule", Line: 1, Column: 7, Context: "    1: iambad\n             ^"})
			test("!@#$%@)#$*", ":root", Declaration{}, "", false, &parse.Error{Message: "unexpected ending in qualified rule", Line: 1, Column: 8, Context: "    1: !@#$%@)#$*\n              ^"})
		})

		Convey("Set dark logo height if not set", func() {
			test(`:root.dark {
  --layout__bg-color: #0047AB;
}
`,
				":root.dark",
				defaultBrandLogoHeight,
				`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`,
				true,
				nil) // appended
		})

		Convey("Do nothing if dark logo height set", func() {
			test(`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 137px;
}
`,
				":root.dark",
				defaultBrandLogoHeight,
				`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 137px;
}
`,
				false,
				nil) // unchanged
		})

		Convey("Set light logo height if not set", func() {
			test(`:root {
  --layout__bg-color: #0047AB;
}
`,
				":root",
				defaultBrandLogoHeight,
				`:root {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`,
				true,
				nil) // appended
		})

		Convey("Do nothing if light logo height set", func() {
			test(`:root {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 137px;
}
`,
				":root",
				defaultBrandLogoHeight,
				`:root {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 137px;
}
`,
				false,
				nil) // unchanged
		})
	})
}
