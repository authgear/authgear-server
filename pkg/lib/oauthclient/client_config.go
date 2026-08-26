package oauthclient

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// ToClientConfig synthesizes a *config.OAuthClientConfig for this dynamic
// client, making it usable everywhere a static-config client is (see
// docs/plans/dcr/2026-08-17-03-client-resolution.md). defaults comes from
// whichever project config governs this row's source — see
// ResolveTokenLifetimes.
func (c *Client) ToClientConfig(defaults *config.OAuthDynamicClientRegistrationDefaultClientConfig) *config.OAuthClientConfig {
	var appType config.OAuthClientApplicationType
	switch {
	case c.Kind == model.OAuthClientKindThirdParty:
		appType = config.OAuthClientApplicationTypeDynamicThirdParty
	case c.ApplicationType == "native":
		appType = config.OAuthClientApplicationTypeNative
	default: // "web", first-party
		appType = config.OAuthClientApplicationTypeSPA
	}

	cfg := &config.OAuthClientConfig{
		ClientID:                       c.ClientID,
		ClientName:                     derefOr(c.ClientName, ""),
		Name:                           c.DisplayName(),
		ApplicationType:                appType,
		RedirectURIs:                   c.RedirectURIs,
		GrantTypes_do_not_use_directly: c.GrantTypes,
		ResponseTypes:                  c.ResponseTypes,
		ClientURI:                      derefOr(c.ClientURI, ""),
		LogoURI:                        derefOr(c.LogoURI, ""),
		TOSURI:                         derefOr(c.TOSURI, ""),
		PolicyURI:                      derefOr(c.PolicyURI, ""),
		IssueJWTAccessToken:            false, // fixed per client.md's "All Authgear extension fields are fixed at their zero values for DCR clients" rule
		IsDynamic:                      true,  // every client built by ToClientConfig is DCR/CIMD-resolved, regardless of Kind
	}
	// nil here means this source has no default_client_config concept at all
	// (see ResolveTokenLifetimes's default case, e.g. a not-yet-implemented
	// CIMD), not "the admin configured no override" -- for DCR, defaults is
	// always non-nil with real token-lifetime values once config defaults
	// have run (config.OAuthDynamicClientRegistrationDefaultClientConfig's
	// own SetDefaults()).
	if defaults != nil {
		cfg.AccessTokenLifetime = defaults.AccessTokenLifetime
		cfg.RefreshTokenLifetime = defaults.RefreshTokenLifetime
		cfg.RefreshTokenIdleTimeoutEnabled = defaults.RefreshTokenIdleTimeoutEnabled
		cfg.RefreshTokenIdleTimeout = defaults.RefreshTokenIdleTimeout
	}
	cfg.SetDefaults() // reuse the exact same fallback logic static clients get
	return cfg
}

// ResolveTokenLifetimes maps a dynamic client's source to the project
// config key that governs its token lifetimes, so no caller has to
// remember which key applies to which source. DCR reads
// oauth.dynamic_client_registration.default_client_config; CIMD will add
// its own case when cimd.md ships.
func ResolveTokenLifetimes(oauthConfig *config.OAuthConfig, source model.OAuthClientSource) *config.OAuthDynamicClientRegistrationDefaultClientConfig {
	switch source {
	case model.OAuthClientSourceDCR:
		return oauthConfig.DynamicClientRegistration.GetDefaultClientConfig()
	default:
		return nil
	}
}
