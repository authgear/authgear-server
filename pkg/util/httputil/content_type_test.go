package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestCheckContentType(t *testing.T) {
	Convey("CheckContentType bad content type", t, func() {
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("good"))
		})

		middleware := httputil.CheckContentType([]string{"application/json"})

		handler := middleware.Handle(originalHandler)

		r, _ := http.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		So(w.Result().StatusCode, ShouldEqual, http.StatusUnsupportedMediaType)
	})

	Convey("CheckContentType good content type", t, func() {
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("good"))
		})

		middleware := httputil.CheckContentType([]string{"application/json"})

		handler := middleware.Handle(originalHandler)

		r, _ := http.NewRequest("POST", "/", nil)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldEqual, "good")
	})
}
