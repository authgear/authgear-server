package jsonpointerutil

import (
	"testing"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAssignToJSONObject(t *testing.T) {
	Convey("TestAssignToJSONObject", t, func() {
		test := func(ptrStr string, value interface{}, expected interface{}) {
			ptr := jsonpointer.MustParse(ptrStr)
			target := make(map[string]interface{})
			err := AssignToJSONObject(ptr, target, value)
			So(err, ShouldBeNil)
			So(target, ShouldResemble, expected)
		}

		test("/a", 42, map[string]interface{}{
			"a": 42,
		})
		test("/a/b", 42, map[string]interface{}{
			"a": map[string]interface{}{
				"b": 42,
			},
		})
	})

	Convey("TestAssignToJSONObject iteratively", t, func() {
		test := func(m map[string]string, expected interface{}) {
			target := make(map[string]interface{})
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
		}, map[string]interface{}{
			"name":        "John Doe",
			"given_name":  "John",
			"family_name": "",
			"address": map[string]interface{}{
				"street_address": "Some street",
				"country":        "HK",
			},
		})
	})
}

func TestRemoveFromJSONObject(t *testing.T) {
	Convey("RemoveFromJSONObject", t, func() {
		test := func(ptrStr string, original interface{}, expected interface{}) {
			ptr := jsonpointer.MustParse(ptrStr)
			err := RemoveFromJSONObject(ptr, original)
			So(err, ShouldBeNil)
			So(original, ShouldResemble, expected)
		}

		test("/a", make(map[string]interface{}), map[string]interface{}{})
		test("/a", map[string]interface{}{
			"a": 42,
		}, map[string]interface{}{})
		test("/a", map[string]interface{}{
			"a": 42,
			"b": "foobar",
		}, map[string]interface{}{
			"b": "foobar",
		})

		test("/a/b", map[string]interface{}{
			"a": map[string]interface{}{
				"b": 42,
			},
		}, map[string]interface{}{
			"a": map[string]interface{}{},
		})
	})
}
