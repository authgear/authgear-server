package validation

import (
	"bytes"
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

		var goValue interface{}

		data := `
		{
			"a": "bad"
		}
		`
		err = validator.ParseReader("#A", bytes.NewReader([]byte(data)), &goValue)
		So(err, ShouldBeError)

		data = `
		{
			"a": 1
		}
		`

		err = validator.ParseReader("#A", bytes.NewReader([]byte(data)), &goValue)
		So(err, ShouldBeNil)
		So(goValue, ShouldResemble, map[string]interface{}{
			"a": 1.0,
		})
	})
}
