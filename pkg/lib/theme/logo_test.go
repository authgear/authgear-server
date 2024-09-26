package theme

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateSetDefaultLogoHeight(t *testing.T) {
	Convey("MigrateSetDefaultLogoHeight", t, func() {
		test := func(input string, expected string) {
			r := strings.NewReader(input)
			result, err := MigrateSetDefaultLogoHeight(r)
			So(err, ShouldBeNil)
			So(string(result), ShouldEqual, expected)
		}

		Convey("Handle empty string", func() {
			test("", "")
		})

		Convey("Set dark logo height if not set", func() {
			test(`:root.dark {
  --layout__bg-color: #0047AB;
}
`, `:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`) // appended
		})

		Convey("Do nothing if dark logo height set", func() {
			Convey("Set dark logo height if not set", func() {
				test(`:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`, `:root.dark {
  --layout__bg-color: #0047AB;
  --brand-logo__height: 40px;
}
`) // unchanged
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
`) // appended
		})

		Convey("Do nothing if light logo height set", func() {
			test(`:root {
  --layout__bg-color: #F0FFFF;
  --brand-logo__height: 40px;
}
`, `:root {
  --layout__bg-color: #F0FFFF;
  --brand-logo__height: 40px;
}
`) // appended
		})
	})
}
