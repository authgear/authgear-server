package config

var _ = FeatureConfigSchema.Add("OAuthFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client": { "$ref": "#/$defs/OAuthClientFeatureConfig" },
		"client_id_metadata_document": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentFeatureConfig" }
	}
}
`)

type OAuthFeatureConfig struct {
	Client                   *OAuthClientFeatureConfig                   `json:"client,omitempty"`
	ClientIDMetadataDocument *OAuthClientIDMetadataDocumentFeatureConfig `json:"client_id_metadata_document,omitempty"`
}

var _ MergeableFeatureConfig = &OAuthFeatureConfig{}

func (c *OAuthFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.OAuth == nil {
		return c
	}

	var merged *OAuthFeatureConfig = c
	if merged == nil {
		merged = &OAuthFeatureConfig{}
	}

	merged.Client = merged.Client.Merge(layer.OAuth.Client)
	merged.ClientIDMetadataDocument = merged.ClientIDMetadataDocument.Merge(layer.OAuth.ClientIDMetadataDocument)

	return merged
}

func (c *OAuthFeatureConfig) GetClientIDMetadataDocument() *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil {
		return nil
	}
	return c.ClientIDMetadataDocument
}

var _ = FeatureConfigSchema.Add("OAuthClientFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"soft_maximum": { "type": "integer" },
		"custom_ui_enabled": { "type": "boolean" },
		"app2app_enabled": { "type": "boolean" }
	}
}
`)

type OAuthClientFeatureConfig struct {
	Maximum         *int  `json:"maximum,omitempty"`
	SoftMaximum     *int  `json:"soft_maximum,omitempty"`
	CustomUIEnabled *bool `json:"custom_ui_enabled,omitempty"`
	App2AppEnabled  *bool `json:"app2app_enabled,omitempty"`
}

func (c *OAuthClientFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = new(99)
	}

	if c.SoftMaximum == nil {
		c.SoftMaximum = new(99)
	}

	if c.CustomUIEnabled == nil {
		c.CustomUIEnabled = new(false)
	}

	if c.App2AppEnabled == nil {
		c.App2AppEnabled = new(false)
	}
}

func (c *OAuthClientFeatureConfig) Merge(layer *OAuthClientFeatureConfig) *OAuthClientFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Maximum != nil {
		c.Maximum = layer.Maximum
	}
	if layer.SoftMaximum != nil {
		c.SoftMaximum = layer.SoftMaximum
	}
	if layer.App2AppEnabled != nil {
		c.App2AppEnabled = layer.App2AppEnabled
	}
	if layer.CustomUIEnabled != nil {
		c.CustomUIEnabled = layer.CustomUIEnabled
	}
	return c
}

var _ = FeatureConfigSchema.Add("OAuthClientIDMetadataDocumentFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"insecure_http_allowed": { "type": "boolean" },
		"insecure_fetch_address_allowed": { "type": "boolean" },
		"rate_limits": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentRateLimitsFeatureConfig" }
	}
}
`)

// OAuthClientIDMetadataDocumentFeatureConfig defeats the CIMD fetch's
// transport protections, for test and local-development projects only.
// Never set on a project serving real traffic, and never at the cluster or
// plan layer -- app-specific override only. See docs/specs/cimd.md § SSRF
// Protection.
type OAuthClientIDMetadataDocumentFeatureConfig struct {
	// InsecureHTTPAllowed permits http:// wherever CIMD requires https: the
	// client_id, the document's uri fields, and the logo fetch. Scheme only.
	InsecureHTTPAllowed *bool `json:"insecure_http_allowed,omitempty"`
	// InsecureFetchAddressAllowed permits connecting to a
	// non-publicly-routable address, including 169.254.169.254.
	InsecureFetchAddressAllowed *bool `json:"insecure_fetch_address_allowed,omitempty"`
	// RateLimits bounds the CIMD fetch path's DoS exposure (docs/specs/cimd.md
	// § Denial of Service). Configurable per plan tier, since the operator --
	// not the tenant -- owns the egress reputation an outbound fetch to an
	// attacker-chosen host spends.
	RateLimits *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig `json:"rate_limits,omitempty"`
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) SetDefaults() {
	if c.InsecureHTTPAllowed == nil {
		c.InsecureHTTPAllowed = new(false)
	}
	if c.InsecureFetchAddressAllowed == nil {
		c.InsecureFetchAddressAllowed = new(false)
	}
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureHTTPAllowed() bool {
	return c != nil && c.InsecureHTTPAllowed != nil && *c.InsecureHTTPAllowed
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureFetchAddressAllowed() bool {
	return c != nil && c.InsecureFetchAddressAllowed != nil && *c.InsecureFetchAddressAllowed
}

// GetRateLimits is nil-safe for the same pre-SetFieldDefaults reason as
// IsInsecureHTTPAllowed. Once defaults have run, RateLimits is always
// non-nil with real bucket values (its own SetDefaults() below).
func (c *OAuthClientIDMetadataDocumentFeatureConfig) GetRateLimits() *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig {
	if c == nil {
		return nil
	}
	return c.RateLimits
}

// Merge is field-level, not a whole-section replace: a higher layer setting
// only one of the two flags (or rate_limits alone) must not reset the
// others back to their code default from a lower layer. *bool (not bool)
// is what lets nil mean "this layer said nothing" -- the only way a higher
// layer can force a flag back off, and the only shape that survives the
// marshal-then-reparse round trip in
// AuthgearFeatureYAMLDescriptor.viewEffectiveResource.
func (c *OAuthClientIDMetadataDocumentFeatureConfig) Merge(layer *OAuthClientIDMetadataDocumentFeatureConfig) *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.InsecureHTTPAllowed != nil {
		c.InsecureHTTPAllowed = layer.InsecureHTTPAllowed
	}
	if layer.InsecureFetchAddressAllowed != nil {
		c.InsecureFetchAddressAllowed = layer.InsecureFetchAddressAllowed
	}
	c.RateLimits = c.RateLimits.Merge(layer.RateLimits)
	return c
}

var _ = FeatureConfigSchema.Add("OAuthClientIDMetadataDocumentRateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"fetch": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig" }
	}
}
`)

// OAuthClientIDMetadataDocumentRateLimitsFeatureConfig is a one-field
// wrapper today, but Fetch is nested under its own "fetch" key -- rather
// than PerProject/PerIP living directly here -- so the JSON path matches
// ratelimit.RateLimitGroupOAuthClientIDMetadataDocumentFetch's dotted name
// ("oauth.client_id_metadata_document.fetch.per_ip") key-for-key, and so a
// sibling action under this same feature (there is none today) has
// somewhere to go without renaming this type again.
type OAuthClientIDMetadataDocumentRateLimitsFeatureConfig struct {
	Fetch *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig `json:"fetch,omitempty"`
}

// GetFetch is nil-safe for the same pre-SetFieldDefaults reason as
// OAuthClientIDMetadataDocumentFeatureConfig.GetRateLimits above. Once
// defaults have run, Fetch is always non-nil (its own SetDefaults()
// below).
func (c *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig) GetFetch() *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig {
	if c == nil {
		return nil
	}
	return c.Fetch
}

// Merge is field-level even though there is only one field today: Fetch
// itself has two real siblings (PerProject/PerIP), so the cascade has to
// reach them -- the same reasoning as the Authenticator -> Password ->
// Policy cascade in feature_authenticator.go, not a shortcut back to a
// whole-section replace.
func (c *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig) Merge(layer *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig) *OAuthClientIDMetadataDocumentRateLimitsFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.Fetch = c.Fetch.Merge(layer.Fetch)
	return c
}

var _ = FeatureConfigSchema.Add("OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_project": { "$ref": "#/$defs/RateLimitConfig" },
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

// OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig bounds
// docs/specs/cimd.md § Denial of Service's two buckets: per project
// (app_id) and per (project, caller IP). There is deliberately no
// per-(project, host) bucket -- see the spec section and Part 4's plan §1.1
// for why.
type OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig struct {
	PerProject *RateLimitConfig `json:"per_project,omitempty"`
	PerIP      *RateLimitConfig `json:"per_ip,omitempty"`
}

// SetDefaults mirrors OAuthDynamicClientRegistrationRateLimitsConfig's
// pattern: PerProject/PerIP are already non-nil by the time this runs
// (SetFieldDefaults force-allocates them), so checking Enabled == nil
// safely detects "the project never configured this bucket" and replaces
// the whole zero-valued struct with the built-in default. Per-IP is
// deliberately tighter than per-project (5/min vs 10/min): the buckets
// count fetches, not requests, and a fetch is deduplicated per client_id
// and per refetch interval, so legitimate per-IP volume is near zero
// regardless of how many users sit behind one NAT (docs/specs/cimd.md §
// Denial of Service).
func (c *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig) SetDefaults() {
	if c.PerProject.Enabled == nil {
		c.PerProject = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   10,
		}
	}
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   5,
		}
	}
}

// Merge replaces each bucket wholesale, not field-by-field:
// enabled/period/burst are one unit, and merging them field-wise would let
// two layers jointly produce a bucket neither one actually wrote.
func (c *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig) Merge(layer *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig) *OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.PerProject != nil {
		c.PerProject = layer.PerProject
	}
	if layer.PerIP != nil {
		c.PerIP = layer.PerIP
	}
	return c
}
