package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCountryCallingCodes(t *testing.T) {
	Convey("CountryCallingCodes", t, func() {
		Convey("No duplicates", func() {
			seen := map[string]struct{}{}
			var duplicates []string
			for _, ccc := range CountryCallingCodes {
				_, ok := seen[ccc]
				if ok {
					duplicates = append(duplicates, ccc)
				}
				seen[ccc] = struct{}{}
			}
			So(duplicates, ShouldBeEmpty)
		})
	})
}
