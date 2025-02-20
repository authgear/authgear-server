package otelauthgear

import (
	"net/http"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		statusCodeAttr := otelutil.HTTPResponseStatusCode(metrics)

		// FIXME: only record if http.route is defined.
		// This essentially makes metric opt-in.
		// For example, we WILL NOT record metric for /healthz.
		_, _ = otelhttp.LabelerFromContext(ctx)

		options := []MetricOption{
			metricOptionAttributeKeyValue{methodAttr},
			metricOptionAttributeKeyValue{schemeAttr},
			metricOptionAttributeKeyValue{statusCodeAttr},
		}

		seconds := metrics.Duration.Seconds()
		Float64HistogramRecord(ctx, HTTPServerRequestDurationHistogram, seconds, options...)
	})
}
