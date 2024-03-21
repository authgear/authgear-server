package slice

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMap(t *testing.T) {
	Convey("Map", t, func() {
		original := []int{1, 2, 3}
		var output []string
		output = Map(original, func(i int) string { return fmt.Sprintf("%d", i) })

		expectedOutput := []string{"1", "2", "3"}
		So(output, ShouldResemble, expectedOutput)
	})
}
