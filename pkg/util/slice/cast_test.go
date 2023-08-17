package slice

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCast(t *testing.T) {
	Convey("Cast", t, func() {
		original := []string{"a", "b", "c"}
		var anySlice []interface{}
		anySlice = Cast[string, interface{}](original)

		stringSlice := Cast[interface{}, string](anySlice)
		So(stringSlice, ShouldResemble, original)
	})
}
