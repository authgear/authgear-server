package graphqlutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLazy(t *testing.T) {
	Convey("Lazy", t, func() {
		must := func(value any, err error) any {
			if err != nil {
				panic(err)
			}
			return value
		}

		Convey("should evaluate value lazily", func() {
			evaluated := false
			lazy := NewLazy(func() (any, error) {
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
			lazy1 := NewLazyValue(func() (any, error) {
				return 123, nil
			})
			So(must(lazy1.Value()), ShouldEqual, 123)

			lazy2 := NewLazyValue(lazy1)
			So(must(lazy2.Value()), ShouldEqual, 123)

			lazy3 := NewLazyValue(123)
			So(must(lazy3.Value()), ShouldEqual, 123)
		})
		Convey("should resolve value recursively", func() {
			lazy := NewLazyValue(func() (any, error) {
				return NewLazyValue(123), nil
			})
			So(must(lazy.Value()), ShouldEqual, 123)
		})
		Convey("should map values", func() {
			eval := 0
			lazy1 := NewLazyValue(func() (any, error) {
				eval++
				return 1, nil
			})
			lazy2 := lazy1.Map(func(i any) (any, error) {
				eval++
				return NewLazy(func() (any, error) {
					eval++
					return i.(int)*10 + 2, nil
				}), nil
			})
			lazy3 := lazy2.Map(func(i any) (any, error) {
				eval++
				return i.(int)*10 + 3, nil
			})

			So(eval, ShouldEqual, 0)
			So(must(lazy3.Value()), ShouldEqual, 123)
			So(eval, ShouldEqual, 4)
		})

		Convey("should resolve value in objects", func() {
			lazy1 := NewLazy(func() (any, error) {
				return map[string]any{
					"key": NewLazyValue(42),
				}, nil
			})
			So(must(lazy1.Value()), ShouldResemble, map[string]any{
				"key": 42,
			})

			lazy2 := NewLazyValue(map[string]any{
				"key1": NewLazyValue(map[string]any{
					"key2": NewLazyValue(42),
				}),
			})
			So(must(lazy2.Value()), ShouldResemble, map[string]any{
				"key1": map[string]any{
					"key2": 42,
				},
			})
		})

		Convey("should resolve value in slices", func() {
			lazy1 := NewLazy(func() (any, error) {
				return []any{
					NewLazyValue(42),
				}, nil
			})
			So(must(lazy1.Value()), ShouldResemble, []any{42})

			lazy2 := NewLazyValue([]any{NewLazyValue([]any{NewLazyValue(42)})})
			So(must(lazy2.Value()), ShouldResemble, []any{[]any{42}})
		})

		Convey("should resolve value in arbitrary JSON compatible structure", func() {
			lazyJSON := NewLazyValue(map[string]any{
				"apps": NewLazyValue([]any{
					NewLazyValue(map[string]any{
						"id": "1",
						"domains": NewLazyValue([]any{
							NewLazyValue(map[string]any{
								"id": "2",
							}),
						}),
					}),
				}),
			})
			So(must(lazyJSON.Value()), ShouldResemble, map[string]any{
				"apps": []any{
					map[string]any{
						"id": "1",
						"domains": []any{
							map[string]any{
								"id": "2",
							},
						},
					},
				},
			})
		})
	})
}
