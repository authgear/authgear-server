package http

import (
	gohttp "net/http"
)

// Header names
const (
	HeaderAPIKey                   = "x-skygear-api-key"
	HeaderAccessToken              = "x-skygear-access-token"
	HeaderAccessKeyType            = "x-skygear-access-key-type"
	HeaderClientID                 = "x-skygear-client-id"
	HeaderUserDisabled             = "x-skygear-user-disabled"
	HeaderUserID                   = "x-skygear-user-userid"
	HeaderUserVerified             = "x-skygear-user-verified"
	HeaderSessionIdentityType      = "x-skygear-session-identity-type"
	HeaderSessionAuthenticatorType = "x-skygear-session-authenticator-type"
	HeaderGear                     = "x-skygear-gear"
	HeaderGearEndpoint             = "x-skygear-gear-endpoint"
	HeaderGearVersion              = "x-skygear-gear-version"
	HeaderHTTPPath                 = "x-skygear-http-path"
	HeaderRequestID                = "x-skygear-request-id"
	HeaderTenantConfig             = "x-skygear-app-config"
	HeaderRequestBodySignature     = "x-skygear-body-signature"
	HeaderSessionExtraInfo         = "x-skygear-extra-info"
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
