package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPermissionsPolicyHeader(t *testing.T) {
	Convey("PermissionsPolicyHeader", t, func() {
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := PermissionsPolicyHeader(dummy)
			return h
		}

		Convey("should set default Permissions-Policy header", func() {
			handler := makeHandler()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			handler.ServeHTTP(recorder, req)

			defaultPermissionsPolicyString := HTTPPermissionsPolicy(DefaultPermissionsPolicy).String()
			So(recorder.Header().Get("Permissions-Policy"), ShouldEqual, defaultPermissionsPolicyString)
		})
	})
}
