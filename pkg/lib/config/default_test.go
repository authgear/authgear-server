package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testStruct struct {
	S  string
	A  *a
	Bs []b
}

func (t *testStruct) SetDefaults() {
	if t.S == "" {
		t.S = "default_string"
	}
}

type a struct {
	Foo *bool
}

func (a *a) SetDefaults() {
	if a.Foo == nil {
		value := false
		a.Foo = &value
	}
}

type b struct {
	Bar float64
}

func (b *b) SetDefaults() {
	if b.Bar == 0 {
		b.Bar = 123
	}
}

func TestSetFieldDefaults(t *testing.T) {
	Convey("SetFieldDefaults", t, func() {
		r := &testStruct{
			Bs: []b{{}},
		}
		SetFieldDefaults(r)

		boolFalse := false
		So(r, ShouldResemble, &testStruct{
			S: "default_string",
			A: &a{
				Foo: &boolFalse,
			},
			Bs: []b{{Bar: 123}},
		})
	})
}
