package otelutil

import (
	"net"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
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

// HTTPServerAddress implements https://opentelemetry.io/docs/specs/semconv/http/http-spans/#setting-serveraddress-and-serverport-attributes
// But the above link does not specify very clearly how to handle IPv6 address.
// Thus, the Go SDK implementation is followed.
// See https://github.com/open-telemetry/opentelemetry-go/blob/main/semconv/internal/v4/net.go#L283
func HTTPServerAddress(hostAndPort string) (*attribute.KeyValue, bool) {
	var hostPart string
	var err error

	if strings.HasPrefix(hostAndPort, "[") {
		addrEnd := strings.LastIndex(hostAndPort, "]")
		if addrEnd < 0 {
			// Invalid input.
			return nil, false
		}

		colon := strings.LastIndex(hostAndPort[addrEnd:], ":")
		if colon < 0 {
			// No port.
		} else {
			// Use SplitHostPort for validation only.
			_, _, err = net.SplitHostPort(hostAndPort)
			if err != nil {
				return nil, false
			}
		}
		hostPart = hostAndPort[1:addrEnd]
	} else {
		colon := strings.LastIndex(hostAndPort, ":")
		if colon < 0 {
			// No port.
			hostPart = hostAndPort
		} else {
			hostPart, _, err = net.SplitHostPort(hostAndPort)
			if err != nil {
				return nil, false
			}
		}
	}

	var addr string
	if parsedIP := net.ParseIP(hostPart); parsedIP != nil {
		addr = parsedIP.String()
	} else {
		addr = hostPart
	}

	out := semconv.ServerAddress(addr)
	return &out, true
}

// HTTPResponseStatusCode implements https://opentelemetry.io/docs/specs/semconv/attributes-registry/http/#http-response-status-code
func HTTPResponseStatusCode(statusCode int) attribute.KeyValue {
	return semconv.HTTPResponseStatusCode(statusCode)
}

func WithHTTPRoute(httpRoute string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			labeler, _ := otelhttp.LabelerFromContext(r.Context())
			labeler.Add(semconv.HTTPRoute(httpRoute))
			next.ServeHTTP(w, r)
		})
	}
}

func WithOtelContext(httpRoute string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, httpRoute)
	}
}

func SetupLabeler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labeler := &otelhttp.Labeler{}
		r = r.WithContext(otelhttp.ContextWithLabeler(r.Context(), labeler))
		next.ServeHTTP(w, r)
	})
}
