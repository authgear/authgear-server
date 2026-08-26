package oauthclient

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// ToModel maps a Client onto the Admin API-facing model.OAuthClient.
// tokenLifetimes should come from ResolveTokenLifetimes(oauthConfig,
// c.Source) — one source of truth for resolved lifetimes: the same
// synthesized config the resolver hands to the OAuth runtime (Part 3).
func (c *Client) ToModel(tokenLifetimes *config.OAuthDynamicClientRegistrationDefaultClientConfig) *model.OAuthClient {
	cfg := c.ToClientConfig(tokenLifetimes)
	applicationType := c.ApplicationType
	return &model.OAuthClient{
		Meta: model.Meta{
			ID:        c.ID, // row uuid — the relay Node id, not ClientID
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt, // == CreatedAt for DCR; real for CIMD refetches
		},
		ClientID:                       c.ClientID,
		Source:                         c.Source,
		Kind:                           c.Kind,
		IsConfidential:                 false,
		IsServiceClient:                false,
		ApplicationType:                &applicationType,
		Name:                           c.DisplayName(),
		ClientName:                     c.ClientName,
		ClientURI:                      c.ClientURI,
		LogoURI:                        c.LogoURI,
		TOSURI:                         c.TOSURI,
		PolicyURI:                      c.PolicyURI,
		RedirectURIs:                   c.RedirectURIs,
		PostLogoutRedirectURIs:         []string{}, // always empty for both dynamic sources
		GrantTypes:                     c.GrantTypes,
		ResponseTypes:                  c.ResponseTypes,
		AccessTokenLifetimeSeconds:     int(cfg.AccessTokenLifetime),
		RefreshTokenLifetimeSeconds:    int(cfg.RefreshTokenLifetime),
		RefreshTokenIdleTimeoutEnabled: *cfg.RefreshTokenIdleTimeoutEnabled, // non-nil after SetDefaults()
		RefreshTokenIdleTimeoutSeconds: int(cfg.RefreshTokenIdleTimeout),
		RefreshTokenRotationEnabled:    false,
		IssueJWTAccessToken:            false, // fixed per client.md's "All Authgear extension fields are fixed at their zero values for DCR clients" rule
		MaxConcurrentSession:           0,
		// all remaining "static clients only" fields (CustomUIURI, App2appEnabled,
		// App2appInsecureDeviceKeyBindingEnabled, DPoPDisabled,
		// PreAuthenticatedURLEnabled, PreAuthenticatedURLAllowedOrigins,
		// ReplaceProjectLogoWithLogoURI) are left at Go zero value, matching
		// client.md's "All Authgear extension fields are fixed at their zero
		// values for DCR clients" rule.
		RegisteredAt:  c.RegisteredAt(), // nil for CIMD, CreatedAt for DCR
		LastFetchedAt: c.LastFetchedAt,  // nil for DCR
	}
}
