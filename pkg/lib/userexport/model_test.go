package userexport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestRequest(t *testing.T) {
	requestBody := `
{
	"format": "ndjson"
}
	`
	Convey("Test serialization of Request", t, func() {
		var request Request
		err := json.Unmarshal([]byte(requestBody), &request)
		So(err, ShouldBeNil)

		serialized, err := json.Marshal(request)
		So(err, ShouldBeNil)

		So(string(serialized), ShouldEqualJSON, `{"csv":{"fields":null},"format":"ndjson"}`)
	})

	Convey("Request JSON Schema", t, func() {
		test := func(requestBody string, errorString string) {
			var request Request
			r, _ := http.NewRequest("POST", "/", strings.NewReader(requestBody))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RequestSchema.Validator(), &request)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}

		test(requestBody, "")
		test("{}", `invalid request body:
<root>: required
  map[actual:<nil> expected:[format] missing:[format]]`)
		test(`
{
	"format": "ndjson"
}
		`, "")
		test(`
{
	"format": "csv"
}
		`, "")
		test(`
{
	"format": "unknown format"
}
		`, `invalid request body:
/format: enum
  map[actual:unknown format expected:[ndjson csv]]`)

		test(`
{
	"format": "csv",
	"csv": {
		"fields": []
	}
}
		`, "")
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "unknown_pointer": "/sub" }]
	}
}
		`, `invalid request body:
/csv/fields/0: required
  map[actual:[unknown_pointer] expected:[pointer] missing:[pointer]]`)
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "pointer": "/sub" }]
	}
}
		`, "")
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "pointer": "/sub", "field_name": "user_id" }]
	}
}
		`, "")
	})
}
