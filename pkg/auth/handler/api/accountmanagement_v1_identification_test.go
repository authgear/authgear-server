package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountManagementV1IdentificationHandlerRequestValidation(t *testing.T) {
	Convey("AccountManagementV1IdentificationHandler request validation", t, func() {
		jsonResponseWriter := httputil.JSONResponseWriter{}
		h := AccountManagementV1IdentificationHandler{
			JSON: &jsonResponseWriter,
		}

		Convey("empty object", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader("{}"))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 400)
			So(w.Body.String(), ShouldEqualJSON, `
{
    "error": {
        "name": "Invalid",
        "reason": "ValidationFailed",
        "message": "invalid request body",
        "code": 400,
        "info": {
            "causes": [
                {
                    "location": "",
                    "kind": "required",
                    "details": {
                        "actual": null,
                        "expected": [
                            "alias",
                            "identification",
                            "redirect_uri"
                        ],
                        "missing": [
                            "alias",
                            "identification",
                            "redirect_uri"
                        ]
                    }
                }
            ]
        }
    }
}
		`)
		})

		Convey("valid", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`
{
	"identification": "oauth",
	"alias": "google",
	"redirect_uri": "myapp.com://host/path"
}
			`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})
}
