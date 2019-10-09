package validation

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidator(t *testing.T) {
	Convey("Validator", t, func() {
		rootSchemaID := "http://example.com"
		schemaString := `
		{
			"$id": "#A",
			"type": "object",
			"properties": {
				"a": { "type": "integer" }
			},
			"required": ["a"]
		}
		`
		validator := NewValidator(rootSchemaID)
		err := validator.AddSchemaFragments(schemaString)
		So(err, ShouldBeNil)

		data := `
		{
			"a": "bad"
		}
		`
		_, err = validator.ValidateReader("#A", bytes.NewReader([]byte(data)))
		So(err, ShouldBeError, `#/a: Invalid type. Expected: integer, given: string
`)

		data = `
		{
			"a": 1
		}
		`

		r, err := validator.ValidateReader("#A", bytes.NewReader([]byte(data)))
		So(err, ShouldBeNil)

		var goValue interface{}
		err = json.NewDecoder(r).Decode(&goValue)
		So(err, ShouldBeNil)
		So(goValue, ShouldResemble, map[string]interface{}{
			"a": 1.0,
		})
	})
}
