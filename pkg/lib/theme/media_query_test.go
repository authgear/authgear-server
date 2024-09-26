package theme

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrateMediaQueryToClassBased(t *testing.T) {
	Convey("MigrateMediaQueryToClassBased", t, func() {
		test := func(input string, expected string) {
			r := strings.NewReader(input)
			result, err := MigrateMediaQueryToClassBased(r)
			So(err, ShouldBeNil)
			So(string(result), ShouldEqual, expected)
		}

		Convey("Handle empty string", func() {
			test("", "")
		})

		Convey("Migrate media query to class-based", func() {
			test(`@media (prefers-color-scheme: dark) {
    :root {
        --color-primary-unshaded: #874bff;
        --color-primary-shaded-1: #faf8ff;
        --color-primary-shaded-2: #ece2ff;
        --color-primary-shaded-3: #dbc9ff;
        --color-primary-shaded-4: #b792ff;
        --color-primary-shaded-5: #9560ff;
        --color-primary-shaded-6: #7943e6;
        --color-primary-shaded-7: #6638c2;
        --color-primary-shaded-8: #4b298f;
        --color-text-unshaded: #ffffff;
        --color-text-shaded-1: #767676;
        --color-text-shaded-2: #a6a6a6;
        --color-text-shaded-3: #c8c8c8;
        --color-text-shaded-4: #d0d0d0;
        --color-text-shaded-5: #dadada;
        --color-text-shaded-6: #eaeaea;
        --color-text-shaded-7: #f4f4f4;
        --color-text-shaded-8: #f8f8f8;
        --color-background-unshaded: #212121;
        --color-background-shaded-1: #e4e4e4;
        --color-background-shaded-2: #cccccc;
        --color-background-shaded-3: #b4b4b4;
        --color-background-shaded-4: #9b9b9b;
        --color-background-shaded-5: #838383;
        --color-background-shaded-6: #6a6a6a;
        --color-background-shaded-7: #525252;
        --color-background-shaded-8: #3a3a3a
    }
}
`, `:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
`)
		})

		Convey("Migrating twice has no effect", func() {
			test(`:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
`, `:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
`)
		})

		Convey("Migrate mixed style", func() {
			test(`@media (prefers-color-scheme: dark) {
    :root {
        --color-primary-unshaded: #874bff;
        --color-primary-shaded-1: #faf8ff;
        --color-primary-shaded-2: #ece2ff;
        --color-primary-shaded-3: #dbc9ff;
        --color-primary-shaded-4: #b792ff;
        --color-primary-shaded-5: #9560ff;
        --color-primary-shaded-6: #7943e6;
        --color-primary-shaded-7: #6638c2;
        --color-primary-shaded-8: #4b298f;
        --color-text-unshaded: #ffffff;
        --color-text-shaded-1: #767676;
        --color-text-shaded-2: #a6a6a6;
        --color-text-shaded-3: #c8c8c8;
        --color-text-shaded-4: #d0d0d0;
        --color-text-shaded-5: #dadada;
        --color-text-shaded-6: #eaeaea;
        --color-text-shaded-7: #f4f4f4;
        --color-text-shaded-8: #f8f8f8;
        --color-background-unshaded: #212121;
        --color-background-shaded-1: #e4e4e4;
        --color-background-shaded-2: #cccccc;
        --color-background-shaded-3: #b4b4b4;
        --color-background-shaded-4: #9b9b9b;
        --color-background-shaded-5: #838383;
        --color-background-shaded-6: #6a6a6a;
        --color-background-shaded-7: #525252;
        --color-background-shaded-8: #3a3a3a
    }
}

:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
`, `:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
:root.dark {
  --color-primary-unshaded: #874bff;
  --color-primary-shaded-1: #faf8ff;
  --color-primary-shaded-2: #ece2ff;
  --color-primary-shaded-3: #dbc9ff;
  --color-primary-shaded-4: #b792ff;
  --color-primary-shaded-5: #9560ff;
  --color-primary-shaded-6: #7943e6;
  --color-primary-shaded-7: #6638c2;
  --color-primary-shaded-8: #4b298f;
  --color-text-unshaded: #ffffff;
  --color-text-shaded-1: #767676;
  --color-text-shaded-2: #a6a6a6;
  --color-text-shaded-3: #c8c8c8;
  --color-text-shaded-4: #d0d0d0;
  --color-text-shaded-5: #dadada;
  --color-text-shaded-6: #eaeaea;
  --color-text-shaded-7: #f4f4f4;
  --color-text-shaded-8: #f8f8f8;
  --color-background-unshaded: #212121;
  --color-background-shaded-1: #e4e4e4;
  --color-background-shaded-2: #cccccc;
  --color-background-shaded-3: #b4b4b4;
  --color-background-shaded-4: #9b9b9b;
  --color-background-shaded-5: #838383;
  --color-background-shaded-6: #6a6a6a;
  --color-background-shaded-7: #525252;
  --color-background-shaded-8: #3a3a3a;
}
`)

		})
	})
}
