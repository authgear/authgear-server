package oidc

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

func ValidateScopes(client *config.OAuthClientConfig, scopes []string) error {
	allowOfflineAccess := false
	for _, grantType := range client.GrantTypes {
		if grantType == "refresh_token" {
			allowOfflineAccess = true
			break
		}
	}
	hasOIDC := false
	hasDeviceSSO := false
	for _, s := range scopes {
		if !IsScopeAllowed(s) {
			return protocol.NewError("invalid_scope", "specified scope is not allowed")
		}
		if s == "offline_access" && !allowOfflineAccess {
			return protocol.NewError("invalid_scope", "offline access is not allowed for this client")
		}
		if s == oauth.FullAccessScope && !client.HasFullAccessScope() {
			return protocol.NewError("invalid_scope", "full access is not allowed for this client")
		}
		if s == "openid" {
			hasOIDC = true
		}
		if s == oauth.DeviceSSOScope {
			hasDeviceSSO = true
		}
		// TODO(tung): Validate if device_sso is allowed by client config
		if s == oauth.DeviceSSOScope && !client.AppInitiatedSSOToWebEnabled {
			return protocol.NewError("invalid_scope", "device_sso is not allowed for this client")
		}
		if s == oauth.AppInitiatedSSOToWebScope && !hasDeviceSSO {
			return protocol.NewError("invalid_scope", "device_sso must be requested when using app-initiated-sso-to-web")
		}
		if s == oauth.AppInitiatedSSOToWebScope && !client.AppInitiatedSSOToWebEnabled {
			return protocol.NewError("invalid_scope", "app-initiated-sso-to-web is not allowed for this client")
		}
	}
	if !hasOIDC {
		return protocol.NewError("invalid_scope", "must request 'openid' scope")
	}
	return nil
}

var AllowedScopes = []string{
	"openid",
	"offline_access",
	oauth.FullAccessScope,
	oauth.FullUserInfoScope,
	oauth.AppInitiatedSSOToWebScope,
	oauth.DeviceSSOScope,
}

func IsScopeAllowed(scope string) bool {
	for _, s := range AllowedScopes {
		if s == scope {
			return true
		}
	}
	return false
}

func IsScopeAuthorized(authorizedScopes []string, requiredScope string) bool {
	for _, scope := range authorizedScopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}
