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
		if s == oauth.OfflineAccess && !allowOfflineAccess {
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
		if s == oauth.DeviceSSOScope && !client.PreAuthenticatedURLEnabled {
			return protocol.NewError("invalid_scope", "device_sso is not allowed for this client")
		}
		if s == oauth.PreAuthenticatedURLScope && !hasDeviceSSO {
			return protocol.NewError("invalid_scope", "device_sso must be requested when using pre-authenticated url")
		}
		if s == oauth.PreAuthenticatedURLScope && !client.PreAuthenticatedURLEnabled {
			return protocol.NewError("invalid_scope", "pre-authenticated url is not allowed for this client")
		}
	}
	if !hasOIDC {
		return protocol.NewError("invalid_scope", "must request 'openid' scope")
	}
	return nil
}

const (
	// The scope openid must be present.
	// https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest
	ScopeOpenID = "openid"
	// Scope "profile" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeProfile = "profile"
	// Scope "email" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeEmail = "email"
	// Scope "address" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeAddress = "address"
	// Scope "phone" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopePhone = "phone"
)

var AllowedScopes = []string{
	// OAuth 2.0 scopes
	oauth.OfflineAccess,
	oauth.DeviceSSOScope,

	// OIDC scopes.
	ScopeOpenID,
	ScopeProfile,
	ScopeEmail,
	ScopeAddress,
	ScopePhone,

	// Authgear specific scopes.
	oauth.FullAccessScope,
	oauth.FullUserInfoScope,
	oauth.PreAuthenticatedURLScope,
}

func IsScopeAllowed(scope string) bool {
	for _, s := range AllowedScopes {
		if s == scope {
			return true
		}
	}
	return false
}
