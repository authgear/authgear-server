package slice

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFlatMap(t *testing.T) {
	Convey("FlatMap", t, func() {
		original := []string{"1 2", "3 4", "5 6"}
		var output []int
		output = FlatMap(original, func(str string) []int {
			splitted := strings.Split(str, " ")
			ints := []int{}
			for _, s := range splitted {
				i, _ := strconv.Atoi(s)
				ints = append(ints, i)
			}
			return ints
		})

		expectedOutput := []int{1, 2, 3, 4, 5, 6}
		So(output, ShouldResemble, expectedOutput)
	})
}
