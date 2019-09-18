package http

import (
	gohttp "net/http"
)

// Header names
const (
	// Headers appearing in client request
	HeaderAPIKey           = "x-skygear-api-key"
	HeaderAccessToken      = "x-skygear-access-token"
	HeaderSessionExtraInfo = "x-skygear-extra-info"

	// Headers appearing in server response
	// When you add a new header, you must expose it in CORSMiddleware.
	// nolint: gosec
	HeaderTryRefreshToken = "x-skygear-try-refresh-token"

	// Headers appearing in proxied request
	HeaderRequestID = "x-skygear-request-id"

	// Headers appearing in proxied microservice request
	HeaderUserID                         = "x-skygear-user-id"
	HeaderUserDisabled                   = "x-skygear-user-disabled"
	HeaderUserVerified                   = "x-skygear-user-verified"
	HeaderSessionIdentityID              = "x-skygear-session-identity-id"
	HeaderSessionIdentityType            = "x-skygear-session-identity-type"
	HeaderSessionIdentityUpdatedAt       = "x-skygear-session-identity-updated-at"
	HeaderSessionAuthenticatorID         = "x-skygear-session-authenticator-id"
	HeaderSessionAuthenticatorType       = "x-skygear-session-authenticator-type"
	HeaderSessionAuthenticatorOOBChannel = "x-skygear-session-authenticator-oob-channel"
	HeaderSessionAuthenticatorUpdatedAt  = "x-skygear-session-authenticator-updated-at"
	HeaderHTTPPath                       = "x-skygear-http-path"

	// Headers appearing in proxied gear request
	HeaderGear         = "x-skygear-gear"
	HeaderGearEndpoint = "x-skygear-gear-endpoint"
	HeaderGearVersion  = "x-skygear-gear-version"
	HeaderTenantConfig = "x-skygear-app-config"

	// Internal headers
	HeaderAccessKeyType = "x-skygear-access-key-type"
	HeaderClientID      = "x-skygear-client-id"

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
