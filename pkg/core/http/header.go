package http

import (
	gohttp "net/http"
	"strings"
)

// Header names
const (
	// Headers appearing in client request
	HeaderSessionExtraInfo = "x-skygear-extra-info"

	// Headers appearing in server response
	// When you add a new header, you must expose it in CORSMiddleware.
	// nolint: gosec
	HeaderTryRefreshToken = "x-skygear-try-refresh-token"

	// Headers appearing in proxied gear request
	HeaderTenantConfig = "x-skygear-app-config"

	// Outbound webhook request
	HeaderRequestBodySignature = "x-skygear-body-signature"
)

func GetHost(req *gohttp.Request) (host string) {
	host = req.Header.Get("X-Forwarded-Host")
	if host != "" {
		return
	}

	host = req.Host
	if host != "" {
		return
	}

	host = req.URL.Host
	return
}

func GetProto(req *gohttp.Request) (proto string) {
	proto = req.Header.Get("X-Forwarded-Proto")
	if proto != "" {
		return
	}

	proto = req.URL.Scheme
	if proto != "" {
		return
	}

	proto = "http"
	return
}

func SetForwardedHeaders(req *gohttp.Request) {
	req.Header.Set("X-Forwarded-Host", GetHost(req))
	req.Header.Set("X-Forwarded-Proto", GetProto(req))
}

// RemoveSkygearHeader removes all x-skygear-* headers.
func RemoveSkygearHeader(header gohttp.Header) gohttp.Header {
	newHeader := gohttp.Header{}
	for name, values := range header {
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "x-skygear-") {
			continue
		}
		newHeader[name] = values
	}
	return newHeader
}

const httpHeaderAuthorization = "authorization"
const httpAuthzBearerScheme = "bearer"

func parseAuthorizationHeader(r *gohttp.Request) (token string) {
	authorization := strings.SplitN(r.Header.Get(httpHeaderAuthorization), " ", 2)
	if len(authorization) != 2 {
		return
	}

	scheme := authorization[0]
	if strings.ToLower(scheme) != httpAuthzBearerScheme {
		return
	}

	return authorization[1]
}

// GetSessionIdentifier extracts session identifier from r.
// The session identifier is either in cookie or Authorization header.
// Cookie has higher precedence.
func GetSessionIdentifier(r *gohttp.Request) string {
	cookie, err := r.Cookie(CookieNameSession)
	if err == nil {
		return cookie.Value
	}
	return parseAuthorizationHeader(r)
}
