package validation_test

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/validation"
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

		err := schema.Validator().Validate(strings.NewReader(`
		{
		}
`))
		So(err, ShouldBeNil)

		err = schema.Validator().Validate(strings.NewReader(`
		{
			"b": "t",
			"c": [
				4,
				5
			]
		}
`))
		So(err, ShouldResemble, &validation.AggregatedError{
			Message: "invalid value",
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

		apiErr := skyerr.AsAPIError(err)
		So(apiErr, ShouldResemble, &skyerr.APIError{
			Kind:    validation.ValidationFailed,
			Message: "invalid value",
			Code:    400,
			Info: map[string]interface{}{
				"causes": []validation.Error{
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
			},
		})
	})
}
