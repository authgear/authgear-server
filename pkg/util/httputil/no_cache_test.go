package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestNoCache(t *testing.T) {
	Convey("NoCache", t, func() {
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("good"))
		})

		middleware := httputil.NoCache

		handler := middleware(originalHandler)

		allMethods := []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}
		for _, method := range allMethods {
			r, _ := http.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")
			So(w.Header().Get("Pragma"), ShouldEqual, "no-cache")
		}
	})
}
