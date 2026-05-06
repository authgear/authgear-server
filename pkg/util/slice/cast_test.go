package slice

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCast(t *testing.T) {
	Convey("Cast", t, func() {
		original := []string{"a", "b", "c"}
		var anySlice []any
		anySlice = Cast[string, any](original)

		stringSlice := Cast[any, string](anySlice)
		So(stringSlice, ShouldResemble, original)
	})
}
