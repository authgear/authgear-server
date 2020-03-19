package oidc

import "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"

func ValidateScopes(scopes []string) error {
	hasOIDC := false
	for _, s := range scopes {
		if !IsScopeAllowed(s) {
			return protocol.NewError("invalid_scope", "specified scope is not allowed")
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
}

func IsScopeAllowed(scope string) bool {
	for _, s := range AllowedScopes {
		if s == scope {
			return true
		}
	}
	return false
}
