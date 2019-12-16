package reflect

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsRecursivelyZero(t *testing.T) {
	f := IsRecursivelyZero
	Convey("IsRecursivelyZero", t, func() {
		Convey("Primitives", func() {
			var arr [1]int
			So(f(false), ShouldBeTrue)
			So(f(0), ShouldBeTrue)
			So(f(0.0), ShouldBeTrue)
			So(f(""), ShouldBeTrue)
			So(f(arr), ShouldBeTrue)
		})

		Convey("Builtin pointer types", func() {
			var ch int
			var fun func(int)
			var m map[int]struct{}
			var s []int

			So(f(ch), ShouldBeTrue)
			So(f(fun), ShouldBeTrue)
			So(f(m), ShouldBeTrue)
			So(f(s), ShouldBeTrue)
		})

		Convey("Pointers", func() {
			var pInt *int
			zero := 0
			one := 1
			So(f(pInt), ShouldBeTrue)
			So(f(&zero), ShouldBeTrue)
			So(f(&one), ShouldBeFalse)
		})

		Convey("Struct", func() {
			type A struct {
				Int    int
				PtrInt *int
			}

			type B struct {
				A A
			}

			type C struct {
				B B
			}

			zero := 0
			So(f(C{}), ShouldBeTrue)
			So(f(&C{}), ShouldBeTrue)
			So(f(&C{
				B: B{
					A: A{
						Int:    0,
						PtrInt: &zero,
					},
				},
			}), ShouldBeTrue)
			So(f(&C{
				B: B{
					A: A{
						Int:    1,
						PtrInt: &zero,
					},
				},
			}), ShouldBeFalse)
		})

	})
}
