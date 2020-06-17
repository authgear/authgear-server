package validation_test

import (
	"github.com/skygeario/skygear-server/pkg/validation"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaValidate(t *testing.T) {
	Convey("validate schema", t, func() {
		schema := validation.NewMultipartSchema("schemaA")
		schema.Add("schemaA", `
		{
			"type": "object",
			"properties": {
				"b": { "$ref": "#/$defs/schemaB" },
				"c": {
					"type": "array",
					"items": { "$ref": "#/$defs/schemaC" }
				}
			}
		}
`)
		schema.Add("schemaB", `
		{
			"type": "string",
			"minLength": 4
		}
`)
		schema.Add("schemaC", `
		{
			"type": "integer",
			"minimum": 5
		}
`)
		schema.Instantiate()

		err := schema.ValidateReader(strings.NewReader(`
		{
		}
`))
		So(err, ShouldBeNil)

		err = schema.ValidateReader(strings.NewReader(`
		{
			"b": "t",
			"c": [
				4,
				5
			]
		}
`))
		So(err, ShouldResemble, &validation.AggregatedError{
			Errors: []validation.Error{
				{
					Location: "/b",
					Keyword:  "minLength",
					Info: map[string]interface{}{
						"actual":   1.0,
						"expected": 4.0,
					},
				},
				{
					Location: "/c/0",
					Keyword:  "minimum",
					Info: map[string]interface{}{
						"actual":  4.0,
						"minimum": 5.0,
					},
				},
			},
		})
	})
}
