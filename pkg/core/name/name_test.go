package name

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateAppName(t *testing.T) {
	Convey("ValidateAppName", t, func() {
		cases := []struct {
			input    string
			expected bool
		}{
			{"", false},              // at least 1 letter
			{"1", false},             // cannot start with digit
			{"a", true},              // good
			{"a23456789012", true},   // longest possible
			{"a234567890123", false}, // too long
		}
		for _, c := range cases {
			actual := ValidateAppName(c.input) == nil
			So(actual, ShouldEqual, c.expected)
		}
	})
}
