package jsonpointerutil

import (
	"testing"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAssignToJSONObject(t *testing.T) {
	Convey("TestAssignToJSONObject", t, func() {
		test := func(ptrStr string, value any, expected any) {
			ptr := jsonpointer.MustParse(ptrStr)
			target := make(map[string]any)
			err := AssignToJSONObject(ptr, target, value)
			So(err, ShouldBeNil)
			So(target, ShouldResemble, expected)
		}

		test("/a", 42, map[string]any{
			"a": 42,
		})
		test("/a/b", 42, map[string]any{
			"a": map[string]any{
				"b": 42,
			},
		})
	})

	Convey("TestAssignToJSONObject iteratively", t, func() {
		test := func(m map[string]string, expected any) {
			target := make(map[string]any)
			for ptrStr, value := range m {
				ptr := jsonpointer.MustParse(ptrStr)
				err := AssignToJSONObject(ptr, target, value)
				So(err, ShouldBeNil)
			}
			So(target, ShouldResemble, expected)
		}

		test(map[string]string{
			"/name":                   "John Doe",
			"/given_name":             "John",
			"/family_name":            "",
			"/address/street_address": "Some street",
			"/address/country":        "HK",
		}, map[string]any{
			"name":        "John Doe",
			"given_name":  "John",
			"family_name": "",
			"address": map[string]any{
				"street_address": "Some street",
				"country":        "HK",
			},
		})
	})
}

func TestRemoveFromJSONObject(t *testing.T) {
	Convey("RemoveFromJSONObject", t, func() {
		test := func(ptrStr string, original any, expected any) {
			ptr := jsonpointer.MustParse(ptrStr)
			err := RemoveFromJSONObject(ptr, original)
			So(err, ShouldBeNil)
			So(original, ShouldResemble, expected)
		}

		test("/a", make(map[string]any), map[string]any{})
		test("/a", map[string]any{
			"a": 42,
		}, map[string]any{})
		test("/a", map[string]any{
			"a": 42,
			"b": "foobar",
		}, map[string]any{
			"b": "foobar",
		})

		test("/a/b", map[string]any{
			"a": map[string]any{
				"b": 42,
			},
		}, map[string]any{
			"a": map[string]any{},
		})
	})
}
