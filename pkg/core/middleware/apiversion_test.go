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
		m := &APIVersionMiddleware{
			APIVersionName: "api_version",
			MajorVersion:   3,
			MinorVersion:   1,
		}
		router.Use(m.Handle)

		router.HandleFunc("/{api_version}/foobar", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		requestURIs := []string{"/nonsense/foobar", "/v2.0/foobar", "/v3.2/foobar"}

		for _, requestURI := range requestURIs {
			r, _ := http.NewRequest("GET", requestURI, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 404,
					"message": "incompatible API version",
					"name": "NotFound",
					"reason": "IncompatibleAPIVersion"
				}
			}
			`)
		}

		r, _ := http.NewRequest("GET", "/v3.1/foobar", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		So(w.Result().StatusCode, ShouldEqual, 200)
		So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
	})
}
