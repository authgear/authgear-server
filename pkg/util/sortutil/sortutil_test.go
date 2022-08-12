package sortutil

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type object struct {
	A int
	B int
	C int
}

func TestAndThen(t *testing.T) {
	Convey("AndThen", t, func() {
		test := func(input []object, expected []object, less func(i int, j int) bool) {
			sort.SliceStable(input, less)
			So(input, ShouldResemble, expected)
		}

		orderByA := func(s []object) func(i int, j int) bool {
			return func(i int, j int) bool {
				ret := s[i].A < s[j].A
				return ret
			}
		}

		orderByB := func(s []object) func(i int, j int) bool {
			return func(i int, j int) bool {
				ret := s[i].B < s[j].B
				return ret
			}
		}
		orderByC := func(s []object) func(i int, j int) bool {
			return func(i int, j int) bool {
				ret := s[i].C < s[j].C
				return ret
			}
		}

		Convey("A -> B -> C", func() {
			i := []object{
				{
					C: 3,
				},
				{
					C: 2,
				},
				{
					C: 1,
				},
			}
			expected := []object{
				{
					C: 1,
				},
				{
					C: 2,
				},
				{
					C: 3,
				},
			}

			less := LessFunc(orderByA(i)).AndThen(orderByB(i)).AndThen(orderByC(i))
			test(i, expected, less)
		})

		Convey("B -> A -> C", func() {
			i := []object{
				{
					A: 3,
				},
				{
					A: 2,
				},
				{
					A: 1,
				},
			}
			expected := []object{
				{
					A: 1,
				},
				{
					A: 2,
				},
				{
					A: 3,
				},
			}

			less := LessFunc(orderByB(i)).AndThen(orderByA(i)).AndThen(orderByC(i))
			test(i, expected, less)
		})

		Convey("C -> B -> A", func() {
			i := []object{
				{
					B: 3,
				},
				{
					B: 2,
				},
				{
					B: 1,
				},
			}
			expected := []object{
				{
					B: 1,
				},
				{
					B: 2,
				},
				{
					B: 3,
				},
			}

			less := LessFunc(orderByC(i)).AndThen(orderByB(i)).AndThen(orderByA(i))
			test(i, expected, less)
		})
	})
}
