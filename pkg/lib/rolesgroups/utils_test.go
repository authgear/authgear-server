package rolesgroups

import (
	"slices"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestComputeKeyDifference(t *testing.T) {
	Convey("computeKeyDifference", t, func() {
		keysToAdd, keysToRemove := computeKeyDifference([]string{"1", "2", "3"}, []string{"1", "2", "3"})
		slices.Sort(keysToAdd)
		slices.Sort(keysToRemove)
		So(keysToAdd, ShouldResemble, []string(nil))
		So(keysToRemove, ShouldResemble, []string(nil))

		keysToAdd, keysToRemove = computeKeyDifference([]string{"1", "2", "3"}, []string{"1"})
		slices.Sort(keysToAdd)
		slices.Sort(keysToRemove)
		So(keysToAdd, ShouldResemble, []string(nil))
		So(keysToRemove, ShouldResemble, []string{"2", "3"})

		keysToAdd, keysToRemove = computeKeyDifference([]string{"1"}, []string{"1", "2", "3"})
		slices.Sort(keysToAdd)
		slices.Sort(keysToRemove)
		So(keysToAdd, ShouldResemble, []string{"2", "3"})
		So(keysToRemove, ShouldResemble, []string(nil))

		keysToAdd, keysToRemove = computeKeyDifference([]string{"1", "2", "3"}, []string{"4", "5", "6"})
		slices.Sort(keysToAdd)
		slices.Sort(keysToRemove)
		So(keysToAdd, ShouldResemble, []string{"4", "5", "6"})
		So(keysToRemove, ShouldResemble, []string{"1", "2", "3"})
	})
}
