package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestAPIVersionMiddleware(t *testing.T) {
	Convey("APIVersionMiddleware", t, func() {
		router := mux.NewRouter()
		// own API version is v3.1
		m := &APIVersionMiddleware{
			APIVersionName:        "api_version",
			SupportedVersions:     []string{"v3.0", "v3.1"},
			SupportedVersionsJSON: `["v3.0", "v3.1"]`,
		}
		router.Use(m.Handle)

		router.HandleFunc("/{api_version}/foobar", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		requestURIs := []string{
			// nonsense is not in API version format.
			"/nonsense/foobar",
			// v2.0 is incompatible because major version mismatches.
			"/v2.0/foobar",
			// v3.2 is incompatible because minor version is newer.
			"/v3.2/foobar",
		}
		for _, requestURI := range requestURIs {
			r, _ := http.NewRequest("GET", requestURI, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 404,
					"message": "expected API versions: [\"v3.0\", \"v3.1\"]",
					"name": "NotFound",
					"reason": "IncompatibleAPIVersion"
				}
			}
			`)
		}

		requestURIs = []string{
			// v3.0 is compatible because major versions are equal and minor version is older.
			"/v3.0/foobar",
			// v3.1 is compatible because it is an exact match.
			"/v3.1/foobar",
		}
		for _, requestURI := range requestURIs {
			r, _ := http.NewRequest("GET", requestURI, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		}
	})
}
