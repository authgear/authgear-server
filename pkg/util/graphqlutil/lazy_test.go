package graphqlutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLazy(t *testing.T) {
	Convey("Lazy", t, func() {
		must := func(value interface{}, err error) interface{} {
			if err != nil {
				panic(err)
			}
			return value
		}

		Convey("should evaluate value lazily", func() {
			evaluated := false
			lazy := NewLazy(func() (interface{}, error) {
				evaluated = true
				return 123, nil
			})

			So(evaluated, ShouldBeFalse)
			value, err := lazy.Value()
			So(err, ShouldBeNil)
			So(value, ShouldEqual, 123)
			So(evaluated, ShouldBeTrue)
		})
		Convey("should construct from values", func() {
			lazy1 := NewLazyValue(func() (interface{}, error) {
				return 123, nil
			})
			So(must(lazy1.Value()), ShouldEqual, 123)

			lazy2 := NewLazyValue(lazy1)
			So(must(lazy2.Value()), ShouldEqual, 123)

			lazy3 := NewLazyValue(123)
			So(must(lazy3.Value()), ShouldEqual, 123)
		})
		Convey("should map values", func() {
			eval := 0
			lazy1 := NewLazyValue(func() (interface{}, error) {
				eval++
				return 1, nil
			})
			lazy2 := lazy1.Map(func(i interface{}) (interface{}, error) {
				eval++
				return NewLazy(func() (interface{}, error) {
					eval++
					return i.(int)*10 + 2, nil
				}), nil
			})
			lazy3 := lazy2.Map(func(i interface{}) (interface{}, error) {
				eval++
				return i.(int)*10 + 3, nil
			})

			So(eval, ShouldEqual, 0)
			So(must(lazy3.Value()), ShouldEqual, 123)
			So(eval, ShouldEqual, 4)
		})
	})
}
