package oidc

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const FullAccessScope = "https://skygear.io/auth-api/full-access"

func ValidateScopes(client config.OAuthClientConfiguration, scopes []string) error {
	allowOfflineAccess := false
	for _, grantType := range client.GrantTypes() {
		if grantType == "refresh_token" {
			allowOfflineAccess = true
			break
		}
	}
	hasOIDC := false
	for _, s := range scopes {
		if !IsScopeAllowed(s) {
			return protocol.NewError("invalid_scope", "specified scope is not allowed")
		}
		if s == "offline_access" && !allowOfflineAccess {
			return protocol.NewError("invalid_scope", "offline access is not allowed for this client")
		}
		if s == "openid" {
			hasOIDC = true
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
	FullAccessScope,
}

func IsScopeAllowed(scope string) bool {
	for _, s := range AllowedScopes {
		if s == scope {
			return true
		}
	}
	return false
}

func SessionScopes(s auth.AuthSession) []string {
	switch s := s.(type) {
	case *session.IDPSession:
		return []string{FullAccessScope}
	case *oauth.OfflineGrant:
		return s.Scopes
	default:
		panic("oidc: unexpected session type")
	}
}
