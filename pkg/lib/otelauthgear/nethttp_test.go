package otelauthgear

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

func TestLabeler(t *testing.T) {
	Convey("Labeler is mutable", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			SetProjectID(r.Context(), "project")
		}

		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r = r.WithContext(otelhttp.ContextWithLabeler(r.Context(), &otelhttp.Labeler{}))
				next.ServeHTTP(w, r)

				labeler, ok := otelhttp.LabelerFromContext(r.Context())
				So(ok, ShouldBeTrue)
				So(labeler.Get(), ShouldResemble, []attribute.KeyValue{
					attributeKeyProjectID.String("project"),
				})
			})
		}

		r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		w := httptest.NewRecorder()

		h := middleware(http.HandlerFunc(handler))
		h.ServeHTTP(w, r)
	})
}
