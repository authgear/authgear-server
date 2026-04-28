package otelauthgear

import (
	"context"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

type HTTPInstrumentationMiddleware struct {
	TrustProxy config.TrustProxy
}

func serveHTTPWithStatus(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
	onDone func(context.Context, int),
) {
	statusCode := 200
	headerWritten := false
	defer func() {
		recovered := recover()
		if recovered != nil {
			if !headerWritten {
				statusCode = 500
			}
		}
		onDone(ctx, statusCode)
		if recovered != nil {
			panic(recovered)
		}
	}()

	wrappedW := httpsnoop.Wrap(w, httpsnoop.Hooks{
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
	next.ServeHTTP(wrappedW, r.WithContext(ctx))
}

func ServeHTTPWithRequestCountMetric(ctx context.Context, w http.ResponseWriter, r *http.Request, next http.Handler) {
	serveHTTPWithStatus(ctx, w, r, next, func(ctx context.Context, statusCode int) {
		otelutil.IntCounterAddOne(
			ctx,
			CounterHTTPRequestCount,
			WithHTTPStatusCode(statusCode),
		)
	})
}

func (m *HTTPInstrumentationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentionally not calling .UTC() to use monotonic clock.
		startTime := time.Now()

		ctx := r.Context()
		// Assume the labeler has been put into context.
		labeler, _ := otelhttp.LabelerFromContext(ctx)

		// Gather method and scheme before invoking the handler.
		// Avoid the rare case of the handler modify r.Method or r.Header.
		labeler.Add(otelutil.HTTPRequestMethod(r))
		scheme := httputil.GetProto(r, bool(m.TrustProxy))
		labeler.Add(otelutil.HTTPURLScheme(scheme))

		serveHTTPWithStatus(ctx, w, r, next, func(ctx context.Context, statusCode int) {
			// Record the request duration.
			requestDuration := time.Since(startTime)

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
				otelutil.Float64HistogramRecord(ctx, HTTPServerRequestDurationHistogram.Inst(), seconds, options...)
			}
		})
	})
}
