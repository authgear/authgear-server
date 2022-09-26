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
	for _, s := range scopes {
		if !IsScopeAllowed(s) {
			return protocol.NewError("invalid_scope", "specified scope is not allowed")
		}
		if s == "offline_access" && !allowOfflineAccess {
			return protocol.NewError("invalid_scope", "offline access is not allowed for this client")
		}
		if s == oauth.FullAccessScope && client.ClientParty() == config.ClientPartyThird {
			return protocol.NewError("invalid_scope", "third-party clients may not request full user access")
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
	oauth.FullAccessScope,
	oauth.FullUserInfoScope,
}

func IsScopeAllowed(scope string) bool {
	for _, s := range AllowedScopes {
		if s == scope {
			return true
		}
	}
	return false
}
