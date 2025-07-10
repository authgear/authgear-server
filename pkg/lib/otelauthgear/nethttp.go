package otelauthgear

import (
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/semconv/v1.27.0"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

type HTTPInstrumentationMiddleware struct {
	TrustProxy config.TrustProxy
}

func (m *HTTPInstrumentationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentionally not calling .UTC() to use monotonic clock.
		startTime := time.Now()

		// 200 is the default.
		// See the documentation of Write() of https://pkg.go.dev/net/http#ResponseWriter
		statusCode := 200
		headerWritten := false

		ctx := r.Context()
		// Assume the labeler has been put into context.
		labeler, _ := otelhttp.LabelerFromContext(ctx)

		// Gather method and scheme before invoking the handler.
		// Avoid the rare case of the handler modify r.Method or r.Header.
		labeler.Add(otelutil.HTTPRequestMethod(r))
		scheme := httputil.GetProto(r, bool(m.TrustProxy))
		labeler.Add(otelutil.HTTPURLScheme(scheme))

		// Wrap w to capture status code.
		w = httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(f httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					f(code)

					if !(code >= 100 && code <= 199) && !headerWritten {
						statusCode = code
						headerWritten = true
					}
				}
			},
		})

		defer func() {
			// Record the request duration.
			requestDuration := time.Since(startTime)

			r := recover()

			// It was a panic and it was not recovered.
			if r != nil {
				// Status code was not written explicitly.
				// Assume 500.
				if !headerWritten {
					statusCode = 500
				}
			}

			// Prepare attributes that is known after serving the request.
			statusCodeAttr := otelutil.HTTPResponseStatusCode(statusCode)

			// Record the metric.
			httpRouteOK := false
			labeler, _ := otelhttp.LabelerFromContext(ctx)
			labelerAttrs := labeler.Get()
			for _, attr := range labelerAttrs {
				if attr.Key == semconv.HTTPRouteKey {
					httpRouteOK = true
				}
			}
			if httpRouteOK {
				// By default, we do not include server.address because it depends on
				// external input like X-Forwarded-Host, Host
				// If we include server.address, then the attacker can trigger cardinality limits.
				options := []otelutil.MetricOption{
					metricOptionAttributeKeyValue{statusCodeAttr},
				}

				seconds := requestDuration.Seconds()
				otelutil.Float64HistogramRecord(ctx, HTTPServerRequestDurationHistogram, seconds, options...)
			}

			// Re-throw the panic.
			if r != nil {
				panic(r)
			}
		}()

		// Invoke the handler.
		next.ServeHTTP(w, r)
	})
}
