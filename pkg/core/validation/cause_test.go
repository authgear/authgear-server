package validation

import (
	"testing"

	"github.com/xeipuuv/gojsonschema"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCause(t *testing.T) {
	type D map[string]interface{}
	// ignore message
	type C struct {
		Kind    ErrorCauseKind
		Pointer string
		Details D
	}

	validate := func(schema, json string) []C {
		result, _ := gojsonschema.Validate(
			gojsonschema.NewStringLoader(schema),
			gojsonschema.NewStringLoader(json),
		)
		if result.Valid() {
			return nil
		}

		causes := toCauses(result.Errors())
		testCauses := make([]C, len(causes))
		for i, c := range causes {
			testCauses[i] = C{c.Kind, c.Pointer, c.Details}
		}
		return testCauses
	}

	Convey("error conversion", t, func() {
		Convey("required", func() {
			causes := validate(`
			{
				"type": "object",
				"properties": {
					"a": { "type": "string" }
				},
				"required": ["a"]
			}`, `{}`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorRequired, Pointer: "/a"},
			})
		})
		Convey("type", func() {
			causes := validate(`{"type": "string"}`, `{}`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorType, Pointer: "", Details: D{"expected": "string"}},
			})
		})
		Convey("const", func() {
			causes := validate(`{"type": "string", "const": "test"}`, `"bad"`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorConstant, Pointer: "", Details: D{"expected": "test"}},
			})
			causes = validate(`{"type": "number", "const": 999}`, `0`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorConstant, Pointer: "", Details: D{"expected": float64(999)}},
			})
		})
		Convey("enum", func() {
			causes := validate(`{"enum": ["test", 999]}`, `null`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorEnum, Pointer: "", Details: D{"expected": []interface{}{"test", float64(999)}}},
			})
		})
		Convey("complex schema", func() {
			schema := `{"oneOf": [{"type": "number"}, {"type": "string", "const": "test"}]}`
			causes := validate(schema, `true`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorType, Pointer: "", Details: D{"expected": "number"}},
			})
			causes = validate(schema, `"bad"`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorConstant, Pointer: "", Details: D{"expected": "test"}},
			})
		})
		Convey("conditional schema", func() {
			schema := `{"if": {"type": "string"}, "then": {"const": "test"}, "else": {"type": "number"}}`
			causes := validate(schema, `true`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorType, Pointer: "", Details: D{"expected": "number"}},
			})
			causes = validate(schema, `"bad"`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorConstant, Pointer: "", Details: D{"expected": "test"}},
			})
		})
		Convey("property name", func() {
			causes := validate(`{"type": "object", "propertyNames": {"type": "string", "enum": ["a", "b"]}}`, `{"c": 123}`)
			So(causes, ShouldResemble, []C{
				C{Kind: ErrorEnum, Pointer: "/c", Details: D{"expected": []interface{}{"a", "b"}}},
			})
		})
	})
}
