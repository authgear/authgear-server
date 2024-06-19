package slice

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilter(t *testing.T) {
	Convey("Filter", t, func() {
		list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		evenNumber := Filter(list, func(i int) bool { return i%2 == 0 })
		expectedOutput := []int{2, 4, 6, 8, 10}
		So(evenNumber, ShouldResemble, expectedOutput)
	})
}
