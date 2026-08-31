package config

var _ = Schema.Add("OAuthClientIDMetadataDocumentConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"allowed_domains": {
			"type": "array",
			"items": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentAllowedDomain" }
		},
		"client_config": { "$ref": "#/$defs/OAuthDynamicClientTokenLifetimesConfig" }
	}
}
`)

// OAuthClientIDMetadataDocumentAllowedDomain constrains an allowed_domains
// entry to the two shapes actually implemented: an exact hostname, or a
// single leading "*." wildcard. A pattern this schema rejects would
// otherwise be accepted by config validation and then silently never match
// anything -- an admin would believe they had allowlisted a domain while
// every request from it was refused. Failing at config load is the only
// place that mistake is cheap to see.
//
// The regex allows: optional "*." prefix, then one or more dot-separated
// LDH labels. A SINGLE-label host such as "localhost" is deliberately
// ALLOWED: rejecting it would be address policy leaking into shape
// validation -- whether a host is reachable is the address filter's job,
// not the allowlist pattern's. It also has to be allowed for a test or
// local-development project to allowlist its own document host.
var _ = Schema.Add("OAuthClientIDMetadataDocumentAllowedDomain", `
{
	"type": "string",
	"minLength": 1,
	"maxLength": 253,
	"pattern": "^(\\*\\.)?[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$"
}
`)

// OAuthClientIDMetadataDocumentConfig carries no nullable tag, so
// config.SetFieldDefaults force-allocates it (and ClientConfig below)
// whenever the whole section is absent from authgear.yaml -- exactly like
// OAuthDynamicClientRegistrationConfig. The nil-safe accessors below exist
// only for code paths that read a config before SetFieldDefaults has run
// (e.g. a test that yaml.Unmarshals a snippet directly); at runtime the
// receiver is always non-nil.
type OAuthClientIDMetadataDocumentConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// AllowedDomains is empty by default, which means "no domain
	// restriction" -- NOT "no domain allowed". The MCP use case is
	// specifically about accepting arbitrary client domains, so an empty
	// allowlist must be permissive. Enforced in
	// oauthclient.IsCIMDClientIDAllowed.
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	// ClientConfig is deliberately not named "default_client_config" the way
	// DCR's is. There is no per-client override path for CIMD, planned or
	// otherwise: a CIMD row is a system-maintained mirror of externally
	// hosted data, overwritten on every refetch, so it is not a place an
	// admin can attach anything. This is simply *the* config for every CIMD
	// client. See docs/specs/cimd.md § Configuration.
	ClientConfig *OAuthDynamicClientTokenLifetimesConfig `json:"client_config,omitempty"`
}

func (c *OAuthClientIDMetadataDocumentConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

func (c *OAuthClientIDMetadataDocumentConfig) GetAllowedDomains() []string {
	if c == nil {
		return nil
	}
	return c.AllowedDomains
}

// GetClientConfig is nil-safe for the same pre-SetFieldDefaults reason as
// IsEnabled. Once defaults have run, ClientConfig is always non-nil with
// real token-lifetime values (OAuthDynamicClientTokenLifetimesConfig's own
// SetDefaults()) -- it is never used to detect "was an override
// configured", because there are no overrides.
func (c *OAuthClientIDMetadataDocumentConfig) GetClientConfig() *OAuthDynamicClientTokenLifetimesConfig {
	if c == nil {
		return nil
	}
	return c.ClientConfig
}
