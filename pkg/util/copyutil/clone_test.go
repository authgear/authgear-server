package copyutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClone(t *testing.T) {
	Convey("it should clone a json", t, func() {
		m1 := map[string]interface{}{
			"a": "bbb",
			"b": map[string]interface{}{
				"c": 123,
			},
		}

		m2, err := Clone(m1)
		So(err, ShouldBeNil)
		So(m2, ShouldResemble, m1)

		// change m1
		m1["a"] = "zzz"
		delete(m1, "b")

		So(m1, ShouldResemble, map[string]interface{}{"a": "zzz"})
		So(m2, ShouldResemble, map[string]interface{}{
			"a": "bbb",
			"b": map[string]interface{}{
				"c": 123,
			},
		})
	})

	Convey("it should clone a interface pointer", t, func() {
		type Nested struct {
			Field string
		}

		type Foo struct {
			Value *interface{}
		}

		ifacePtr := func(v interface{}) *interface{} {
			return &v
		}

		v := Foo{
			Value: ifacePtr(Nested{Field: "111"}),
		}
		vv, err := Clone(v)
		So(err, ShouldBeNil)
		So(vv, ShouldResemble, v)
	})

	Convey("it should copy primitives", t, func() {
		ps := []interface{}{
			42,
			"foo",
			1.2,
		}

		ppss, err := Clone(ps)
		So(err, ShouldBeNil)

		So(ppss, ShouldResemble, ps)

	})

	Convey("it should copy primitives pointers", t, func() {
		i := 42
		s := "foo"
		f := 1.2
		ps := []interface{}{
			&i,
			&s,
			&f,
		}

		ppss, err := Clone(ps)
		So(err, ShouldBeNil)

		So(ppss, ShouldResemble, ps)

	})
}
