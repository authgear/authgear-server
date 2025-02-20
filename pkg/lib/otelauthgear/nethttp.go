package otelauthgear

import (
	"net/http"

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
		ctx := r.Context()
		ctx = otelhttp.ContextWithLabeler(ctx, &otelhttp.Labeler{})
		r = r.WithContext(ctx)

		methodAttr := otelutil.HTTPRequestMethod(r)
		scheme := httputil.GetProto(r, bool(m.TrustProxy))
		schemeAttr := otelutil.HTTPURLScheme(scheme)

		// Invoke the handler.
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		statusCodeAttr := otelutil.HTTPResponseStatusCode(metrics)

		httpRouteOK := false
		labeler, _ := otelhttp.LabelerFromContext(ctx)
		labelerAttrs := labeler.Get()
		for _, attr := range labelerAttrs {
			if attr.Key == semconv.HTTPRouteKey {
				httpRouteOK = true
			}
		}
		if httpRouteOK {
			options := []MetricOption{
				metricOptionAttributeKeyValue{methodAttr},
				metricOptionAttributeKeyValue{schemeAttr},
				metricOptionAttributeKeyValue{statusCodeAttr},
			}

			seconds := metrics.Duration.Seconds()
			Float64HistogramRecord(ctx, HTTPServerRequestDurationHistogram, seconds, options...)
		}
	})
}
