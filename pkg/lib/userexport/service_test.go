package userexport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseExportRequest(t *testing.T) {
	Convey("ParseExportRequest", t, func() {
		svc := &UserExportService{}

		test := func(requestBody string, errorString string) {
			r, _ := http.NewRequest("POST", "/", strings.NewReader(requestBody))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			_, err := svc.ParseExportRequest(w, r)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}

		test(`{"format":"ndjson"}`, "")
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
		`, `invalid request body:
/csv/fields: minItems
  map[actual:0 expected:1]`)
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
		"fields": [{ "pointer": "invalid" }]
	}
}
		`, `invalid request body:
/csv/fields/0/pointer: format
  map[error:0: expecting / but found: "i" format:json-pointer]`)
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
		"fields": [{ "pointer": "/sub", "field_name": "" }]
	}
}
		`, `invalid request body:
/csv/fields/0/field_name: minLength
  map[actual:0 expected:1]`)
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
