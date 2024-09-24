package setutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSet(t *testing.T) {
	Convey("Set", t, func() {
		slice1 := []string{"1", "3", "5"}
		slice2 := []string{"1", "2", "6"}

		s1 := NewSetFromSlice(slice1, Identity[string])
		s2 := NewSetFromSlice(slice2, Identity[string])

		So(SetToSlice(slice1, s1.Subtract(s2), Identity[string]), ShouldResemble, []string{"3", "5"})
		So(SetToSlice(slice2, s2.Subtract(s1), Identity[string]), ShouldResemble, []string{"2", "6"})
	})

	Convey("Set.Merge", t, func() {
		slice1 := []string{"1", "3", "5"}
		slice2 := []string{"1", "2", "6"}

		s1 := NewSetFromSlice(slice1, Identity[string])
		s2 := NewSetFromSlice(slice2, Identity[string])
		result := s1.Merge(s2)

		So(result.Keys(), ShouldResemble, []string{"1", "2", "3", "5", "6"})
	})

}
