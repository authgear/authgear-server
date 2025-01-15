package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	AuthorizationCodeGrantType = "authorization_code"
	RefreshTokenGrantType      = "refresh_token"
	// nolint:gosec
	TokenExchangeGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"

	AnonymousRequestGrantType = "urn:authgear:params:oauth:grant-type:anonymous-request"
	BiometricRequestGrantType = "urn:authgear:params:oauth:grant-type:biometric-request"
	App2AppRequestGrantType   = "urn:authgear:params:oauth:grant-type:app2app-request"
	// nolint:gosec
	IDTokenGrantType        = "urn:authgear:params:oauth:grant-type:id-token"
	SettingsActionGrantType = "urn:authgear:params:oauth:grant-type:settings-action"
)

// whitelistedGrantTypes is a list of grant types that would be always allowed
// to all clients.
var whitelistedGrantTypes = []string{
	AuthorizationCodeGrantType,
	RefreshTokenGrantType,

	TokenExchangeGrantType,

	AnonymousRequestGrantType,
	BiometricRequestGrantType,
	App2AppRequestGrantType,
	IDTokenGrantType,
	SettingsActionGrantType,
}

func GetAllowedGrantTypes(c *config.OAuthClientConfig) []string {
	seen := make(map[string]struct{})

	var allowed []string
	for _, g := range c.GrantTypes_do_not_use_directly {
		_, ok := seen[g]
		if !ok {
			allowed = append(allowed, g)
			seen[g] = struct{}{}
		}
	}

	for _, g := range whitelistedGrantTypes {
		_, ok := seen[g]
		if !ok {
			allowed = append(allowed, g)
			seen[g] = struct{}{}
		}
	}

	return allowed
}
