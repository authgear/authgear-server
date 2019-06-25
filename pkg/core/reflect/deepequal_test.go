package reflect

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNonRecursiveDataDeepEqual(t *testing.T) {
	type S struct {
		A string
	}

	cases := []struct {
		v1      interface{}
		v2      interface{}
		outcome bool
	}{
		{nil, nil, true},

		{nil, []int{}, true},
		{[]int{}, nil, true},

		{nil, map[int]int{}, true},
		{map[int]int{}, nil, true},

		{"a", "a", true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{map[int]int{1: 1}, map[int]int{1: 1}, true},

		{S{A: "a"}, S{A: "a"}, true},
	}

	Convey("TestNonRecursiveDataDeepEqual", t, func() {
		for _, c := range cases {
			So(NonRecursiveDataDeepEqual(c.v1, c.v2), ShouldEqual, c.outcome)
		}
	})
}
