package otelutil

import (
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

// The following attributes are not supported because we do not know the actual protocol between
// the client and the reverse proxy.
// And they are not very essential.
// https://opentelemetry.io/docs/specs/semconv/attributes-registry/network/#network-protocol-name
// https://opentelemetry.io/docs/specs/semconv/attributes-registry/network/#network-protocol-version

// HTTPRequestMethod implements https://opentelemetry.io/docs/specs/semconv/attributes-registry/http/#http-request-method
func HTTPRequestMethod(r *http.Request) attribute.KeyValue {
	upper := strings.ToUpper(r.Method)
	// Per https://opentelemetry.io/docs/specs/semconv/attributes-registry/http/#http-request-method
	// We need to recognize the methods defined in https://www.rfc-editor.org/rfc/rfc9110.html#name-methods
	// and the PATCH method defined in https://www.rfc-editor.org/rfc/rfc5789.html
	switch upper {
	case "GET":
		return semconv.HTTPRequestMethodGet
	case "HEAD":
		return semconv.HTTPRequestMethodHead
	case "POST":
		return semconv.HTTPRequestMethodPost
	case "PUT":
		return semconv.HTTPRequestMethodPut
	case "DELETE":
		return semconv.HTTPRequestMethodDelete
	case "CONNECT":
		return semconv.HTTPRequestMethodConnect
	case "OPTIONS":
		return semconv.HTTPRequestMethodOptions
	case "TRACE":
		return semconv.HTTPRequestMethodTrace
	case "PATCH":
		return semconv.HTTPRequestMethodPatch
	default:
		return semconv.HTTPRequestMethodOther
	}
}

// HTTPURLScheme implements https://opentelemetry.io/docs/specs/semconv/attributes-registry/url/#url-scheme
func HTTPURLScheme(scheme string) attribute.KeyValue {
	switch scheme {
	case "http":
		return semconv.URLScheme(scheme)
	case "https":
		return semconv.URLScheme(scheme)
	default:
		// Assume http.
		return semconv.URLScheme("http")
	}
}

// HTTPResponseStatusCode implements https://opentelemetry.io/docs/specs/semconv/attributes-registry/http/#http-response-status-code
func HTTPResponseStatusCode(metrics httpsnoop.Metrics) attribute.KeyValue {
	statusCode := metrics.Code
	return semconv.HTTPResponseStatusCode(statusCode)
}
