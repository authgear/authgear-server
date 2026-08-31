package config

var _ = Schema.Add("OAuthDynamicClientRegistrationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"initial_access_token_required": { "type": "boolean" },
		"default_client_config": { "$ref": "#/$defs/OAuthDynamicClientTokenLifetimesConfig" },
		"rate_limits": { "$ref": "#/$defs/OAuthDynamicClientRegistrationRateLimitsConfig" }
	}
}
`)

// OAuthDynamicClientRegistrationConfig has no nullable tag, so
// config.SetFieldDefaults force-allocates it (and, transitively,
// DefaultClientConfig/RateLimits below) whenever the whole section is
// absent from authgear.yaml -- IsEnabled/IsInitialAccessTokenRequired stay
// nil-safe only for code paths that read a config before SetFieldDefaults
// has run (e.g. a test that yaml.Unmarshals a snippet directly); at
// runtime the receiver is always non-nil.
type OAuthDynamicClientRegistrationConfig struct {
	Enabled                    bool                                            `json:"enabled,omitempty"`
	InitialAccessTokenRequired *bool                                           `json:"initial_access_token_required,omitempty"`
	DefaultClientConfig        *OAuthDynamicClientTokenLifetimesConfig         `json:"default_client_config,omitempty"`
	RateLimits                 *OAuthDynamicClientRegistrationRateLimitsConfig `json:"rate_limits,omitempty"`
}

func (c *OAuthDynamicClientRegistrationConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// IsInitialAccessTokenRequired defaults to true (the spec default) when the
// section or the field itself is absent.
func (c *OAuthDynamicClientRegistrationConfig) IsInitialAccessTokenRequired() bool {
	return c == nil || c.InitialAccessTokenRequired == nil || *c.InitialAccessTokenRequired
}

// GetDefaultClientConfig is nil-safe for the same pre-SetFieldDefaults
// reason as IsEnabled above. Once defaults have run, DefaultClientConfig is
// always non-nil with real token-lifetime values (its own SetDefaults()
// below) -- it is never used to detect "was an override configured."
func (c *OAuthDynamicClientRegistrationConfig) GetDefaultClientConfig() *OAuthDynamicClientTokenLifetimesConfig {
	if c == nil {
		return nil
	}
	return c.DefaultClientConfig
}

// GetRateLimits is nil-safe for the same pre-SetFieldDefaults reason as
// IsEnabled above. Once defaults have run, RateLimits (and its PerIP/
// PerProject fields) is always non-nil (its own SetDefaults() below).
func (c *OAuthDynamicClientRegistrationConfig) GetRateLimits() *OAuthDynamicClientRegistrationRateLimitsConfig {
	if c == nil {
		return nil
	}
	return c.RateLimits
}

var _ = Schema.Add("OAuthDynamicClientTokenLifetimesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"access_token_lifetime_seconds": { "type": "integer", "minimum": 1 },
		"refresh_token_lifetime_seconds": { "type": "integer", "minimum": 1 },
		"refresh_token_idle_timeout_enabled": { "type": "boolean" },
		"refresh_token_idle_timeout_seconds": { "type": "integer", "minimum": 1 }
	}
}
`)

// Field names/types/json tags match the corresponding subset of
// OAuthClientConfig exactly, since these values are later copied onto the
// synthetic OAuthClientConfig built for a resolved DCR client (Part 3).
type OAuthDynamicClientTokenLifetimesConfig struct {
	AccessTokenLifetime            DurationSeconds `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime           DurationSeconds `json:"refresh_token_lifetime_seconds,omitempty"`
	RefreshTokenIdleTimeoutEnabled *bool           `json:"refresh_token_idle_timeout_enabled,omitempty"`
	RefreshTokenIdleTimeout        DurationSeconds `json:"refresh_token_idle_timeout_seconds,omitempty"`
}

// SetDefaults mirrors OAuthClientConfig.SetDefaults()'s token-lifetime
// fallback exactly (same constants, same zero-means-unset per-field
// semantics), so a DCR client with no configured override resolves to the
// identical values a static client that omits every lifetime field would
// get. This is deliberately not a copy of OAuthClientConfig.SetDefaults()
// itself: that method also has an ApplicationType-driven branch
// (IssueJWTAccessToken) that does not apply here.
func (c *OAuthDynamicClientTokenLifetimesConfig) SetDefaults() {
	if c.AccessTokenLifetime == 0 {
		c.AccessTokenLifetime = DefaultAccessTokenLifetime
	}
	if c.RefreshTokenLifetime == 0 {
		c.RefreshTokenLifetime = max(c.AccessTokenLifetime, DefaultRefreshTokenLifetime)
	}
	if c.RefreshTokenIdleTimeoutEnabled == nil {
		b := DefaultRefreshTokenIdleTimeoutEnabled
		c.RefreshTokenIdleTimeoutEnabled = &b
	}
	if c.RefreshTokenIdleTimeout == 0 {
		c.RefreshTokenIdleTimeout = DefaultRefreshTokenIdleTimeout
	}
}

var _ = Schema.Add("OAuthDynamicClientRegistrationRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_project": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

// OAuthDynamicClientRegistrationRateLimitsConfig lets a project raise or
// lower POST /oauth2/register's rate limits (docs/specs/dcr.md's Rate
// Limits section) -- e.g. an MCP-style integration where many distinct,
// legitimate installs each self-register once may need a higher
// per_project allowance than the built-in default.
type OAuthDynamicClientRegistrationRateLimitsConfig struct {
	PerIP      *RateLimitConfig `json:"per_ip,omitempty"`
	PerProject *RateLimitConfig `json:"per_project,omitempty"`
}

// SetDefaults mirrors AuthenticationRateLimitsSignupConfig.SetDefaults()'s
// pattern exactly: PerIP/PerProject are already non-nil by the time this
// runs (config.SetFieldDefaults force-allocates them first, since neither
// carries a nullable tag), so checking Enabled == nil safely detects
// "the project never configured this bucket" and replaces the whole
// zero-valued struct with the built-in default.
func (c *OAuthDynamicClientRegistrationRateLimitsConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   10,
		}
	}
	if c.PerProject.Enabled == nil {
		c.PerProject = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1h",
			Burst:   1000,
		}
	}
}
