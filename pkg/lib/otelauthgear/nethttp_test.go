package otelauthgear

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	httpconv "go.opentelemetry.io/otel/semconv/v1.34.0/httpconv"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
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

type mockFloat64Histogram struct {
	embedded.Float64Histogram
	called  bool
	options []metric.RecordOption
}

func (h *mockFloat64Histogram) Enabled(context.Context) bool {
	return true
}

var _ metric.Float64Histogram = (*mockFloat64Histogram)(nil)

func (h *mockFloat64Histogram) Record(ctx context.Context, incr float64, options ...metric.RecordOption) {
	h.called = true
	h.options = options
}

func TestHTTPInstrumentationMiddleware(t *testing.T) {
	Convey("HTTPInstrumentationMiddleware", t, func() {
		mock := &mockFloat64Histogram{}

		original := HTTPServerRequestDurationHistogram
		HTTPServerRequestDurationHistogram = httpconv.ServerRequestDuration{
			Float64Histogram: mock,
		}
		defer func() {
			HTTPServerRequestDurationHistogram = original
		}()

		m := &HTTPInstrumentationMiddleware{}

		test := func(h http.Handler, called bool) {
			w := httptest.NewRecorder()

			r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			r = r.WithContext(otelhttp.ContextWithLabeler(r.Context(), &otelhttp.Labeler{}))

			h = m.Handle(h)
			h.ServeHTTP(w, r)
			So(mock.called, ShouldEqual, called)
		}

		Convey("does not record if http.route is undefined", func() {
			test(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), false)
		})

		Convey("record if http.route is defined", func() {
			m := otelutil.WithHTTPRoute("/myroute")
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})

			test(m(h), true)
		})

		Convey("record if the handler handles the panic", func() {
			m := otelutil.WithHTTPRoute("/myroute")
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if r := recover(); r != nil {
						// recover
					}
				}()
				panic(errors.New("panic"))
			})
			test(m(h), true)
		})

		Convey("record even if the handler does not handle the panic", func() {
			err := errors.New("panic")
			m := otelutil.WithHTTPRoute("/myroute")
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(err)
			})

			So(func() {
				test(m(h), true)
			}, ShouldPanicWith, err)
		})
	})
}

func TestServeHTTPWithStatus(t *testing.T) {
	Convey("serveHTTPWithStatus", t, func() {
		Convey("reports status code written by handler", func() {
			var gotStatus int

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			serveHTTPWithStatus(context.Background(), w, r, handler, func(ctx context.Context, statusCode int) {
				gotStatus = statusCode
			})

			So(gotStatus, ShouldEqual, http.StatusNoContent)
		})

		Convey("reports 500 when panic happens before headers are written and re-panics", func() {
			var gotStatus int
			errPanic := errors.New("panic")

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(errPanic)
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			So(func() {
				serveHTTPWithStatus(context.Background(), w, r, handler, func(ctx context.Context, statusCode int) {
					gotStatus = statusCode
				})
			}, ShouldPanicWith, errPanic)

			So(gotStatus, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
