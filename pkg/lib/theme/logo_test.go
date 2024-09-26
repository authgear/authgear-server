package theme

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateSetDefaultLogoHeight(t *testing.T) {
	Convey("MigrateSetDefaultLogoHeight", t, func() {
		test := func(input string, expected string, expectedAlreadySet bool) {
			r := strings.NewReader(input)
			result, alreadySet, err := MigrateSetDefaultLogoHeight(r)
			So(err, ShouldBeNil)
			So(alreadySet, ShouldEqual, expectedAlreadySet)
			So(string(result), ShouldEqual, expected)
		}

		Convey("Handle empty string", func() {
			test("", "", false)
		})

		Convey("Set dark logo height if not set", func() {
			test(`:root.dark {
  --layout__bg-color: #0047AB;
}
`, `:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`, false) // appended
		})

		Convey("Do nothing if dark logo height set", func() {
			Convey("Set dark logo height if not set", func() {
				test(`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`, ``, true) // unchanged
			})
		})

		Convey("Set light logo height if not set", func() {
			test(`:root {
  --layout__bg-color: #F0FFFF;
}
`, `:root {
  --layout__bg-color: #F0FFFF;
  --brand-logo__height: 40px;
}
`, false) // appended
		})

		Convey("Do nothing if light logo height set", func() {
			test(`:root {
  --layout__bg-color: #F0FFFF;
  --brand-logo__height: 40px;
}
`, ``, true) // appended
		})
	})
}
