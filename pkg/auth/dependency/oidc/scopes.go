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

func IsScopeAllowed(scope string) bool {
	switch scope {
	case "openid",
		"offline_access":
		return true
	}
	return false
}
