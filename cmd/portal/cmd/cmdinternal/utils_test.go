package cmdinternal

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMap(t *testing.T) {
	fromJSON := func(text string) map[string]any {
		var value map[string]any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			panic(err)
		}
		return value
	}
	toJSON := func(value any) string {
		text, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return string(text)
	}

	Convey("mapGet", t, func() {
		Convey("get value using key path", func() {
			value := fromJSON(`
			{ "a": {"b": {"c": null, "d": 123}, "e": "text" }, "f": [] }
			`)

			var r any

			_, ok := mapGet[any](value, "a", "b", "c")
			So(ok, ShouldBeFalse)

			r, ok = mapGet[float64](value, "a", "b", "d")
			So(ok, ShouldBeTrue)
			So(r, ShouldResemble, float64(123))

			_, ok = mapGet[string](value, "a", "b", "d")
			So(ok, ShouldBeFalse)

			r, ok = mapGet[map[string]any](value, "a", "b")
			So(ok, ShouldBeTrue)
			So(r, ShouldResemble, map[string]any{"c": nil, "d": float64(123)})

			r, ok = mapGet[[]any](value, "f")
			So(ok, ShouldBeTrue)
			So(r, ShouldResemble, []any{})

			r, ok = mapGet[map[string]any](value)
			So(ok, ShouldBeTrue)
			So(r, ShouldResemble, value)
		})
	})

	Convey("mapSet", t, func() {
		Convey("get value using key path", func() {
			m := make(map[string]any)

			mapSet(m, true, "a")
			So(toJSON(m), ShouldEqualJSON, `{"a": true}`)

			mapSet(m, "test", "b", "c")
			So(toJSON(m), ShouldEqualJSON, `{"a": true, "b": {"c": "test"}}`)

			mapSet(m, 123, "b", "d")
			So(toJSON(m), ShouldEqualJSON, `{"a": true, "b": {"c": "test", "d": 123}}`)
		})
	})
}
