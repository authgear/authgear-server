package config

var _ = FeatureConfigSchema.Add("AdminAPIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"create_session_enabled": { "type": "boolean" },
		"user_import_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"user_export_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"rate_limits": { "$ref": "#/$defs/AdminAPIRateLimitsFeatureConfig" }
	}
}
`)

type AdminAPIFeatureConfig struct {
	CreateSessionEnabled *bool `json:"create_session_enabled,omitempty"`
	// UserImportUsage is the usage limit on user import API, measured by number of imported users.
	UserImportUsage *Deprecated_UsageLimitConfig `json:"user_import_usage,omitempty"`
	// UserExportUsage is the usage limit on user export API, measured by number of export requests.
	UserExportUsage *Deprecated_UsageLimitConfig `json:"user_export_usage,omitempty"`
	// RateLimits bounds Admin API request volume (docs/specs/rate-limit.md
	// § Admin API mutations). It lives in feature config with no authgear.yaml
	// counterpart on purpose: the limit bounds what a project's own
	// administrators can do to that project's storage, so letting the same
	// party raise it would defeat it.
	RateLimits *AdminAPIRateLimitsFeatureConfig `json:"rate_limits,omitempty"`
}

// GetRateLimits is nil-safe for callers that read a config before
// SetFieldDefaults has run (e.g. a test that unmarshals a YAML snippet
// directly). At runtime the whole chain is non-nil, because SetFieldDefaults
// force-allocates every pointer without a nullable tag.
func (c *AdminAPIFeatureConfig) GetRateLimits() *AdminAPIRateLimitsFeatureConfig {
	if c == nil {
		return nil
	}
	return c.RateLimits
}

var _ MergeableFeatureConfig = &AdminAPIFeatureConfig{}

func (c *AdminAPIFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.AdminAPI == nil {
		return c
	}

	var merged *AdminAPIFeatureConfig = c
	if merged == nil {
		merged = &AdminAPIFeatureConfig{}
	}

	if layer.AdminAPI.CreateSessionEnabled != nil {
		merged.CreateSessionEnabled = layer.AdminAPI.CreateSessionEnabled
	}

	if layer.AdminAPI.UserImportUsage != nil {
		merged.UserImportUsage = layer.AdminAPI.UserImportUsage
	}

	if layer.AdminAPI.UserExportUsage != nil {
		merged.UserExportUsage = layer.AdminAPI.UserExportUsage
	}

	merged.RateLimits = merged.RateLimits.Merge(layer.AdminAPI.RateLimits)

	return merged
}

func (c *AdminAPIFeatureConfig) SetDefaults() {
	if c.CreateSessionEnabled == nil {
		c.CreateSessionEnabled = new(false)
	}
	if c.UserImportUsage.Enabled == nil {
		c.UserImportUsage = &Deprecated_UsageLimitConfig{
			Enabled: new(false),
		}
	}
	if c.UserExportUsage.Enabled == nil {
		c.UserExportUsage = &Deprecated_UsageLimitConfig{
			Enabled: new(false),
		}
	}
}

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"mutation": { "$ref": "#/$defs/AdminAPIRateLimitsMutationFeatureConfig" }
	}
}
`)

// AdminAPIRateLimitsFeatureConfig has one field today. Mutation is nested under
// its own key -- rather than the buckets living directly here -- so the JSON
// path matches ratelimit.RateLimitGroupAdminAPIMutation's dotted name
// ("admin_api.mutation.all.per_ip") key for key, and so a sibling action under
// this same section has somewhere to go without renaming this type again.
type AdminAPIRateLimitsFeatureConfig struct {
	Mutation *AdminAPIRateLimitsMutationFeatureConfig `json:"mutation,omitempty"`
}

// GetMutation is nil-safe for the same pre-SetFieldDefaults reason as
// AdminAPIFeatureConfig.GetRateLimits.
func (c *AdminAPIRateLimitsFeatureConfig) GetMutation() *AdminAPIRateLimitsMutationFeatureConfig {
	if c == nil {
		return nil
	}
	return c.Mutation
}

// Merge is field-level even with a single field today, because the cascade has
// to reach All's real siblings further down -- the same reasoning as the
// Authenticator -> Password -> Policy cascade in feature_authenticator.go.
func (c *AdminAPIRateLimitsFeatureConfig) Merge(layer *AdminAPIRateLimitsFeatureConfig) *AdminAPIRateLimitsFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.Mutation = c.Mutation.Merge(layer.Mutation)
	return c
}

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsMutationFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"all": { "$ref": "#/$defs/AdminAPIRateLimitsMutationScopeFeatureConfig" }
	}
}
`)

// AdminAPIRateLimitsMutationFeatureConfig is keyed by scope. "all" is the
// reserved scope covering every mutation field; scopes for individual mutations
// may be added as siblings of it later, keyed by the mutation's field name in
// snake_case, and would be consumed in addition to "all" rather than as a
// fallback.
type AdminAPIRateLimitsMutationFeatureConfig struct {
	All *AdminAPIRateLimitsMutationScopeFeatureConfig `json:"all,omitempty"`
}

// GetAll is nil-safe for the same pre-SetFieldDefaults reason as
// AdminAPIFeatureConfig.GetRateLimits.
func (c *AdminAPIRateLimitsMutationFeatureConfig) GetAll() *AdminAPIRateLimitsMutationScopeFeatureConfig {
	if c == nil {
		return nil
	}
	return c.All
}

func (c *AdminAPIRateLimitsMutationFeatureConfig) Merge(layer *AdminAPIRateLimitsMutationFeatureConfig) *AdminAPIRateLimitsMutationFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.All = c.All.Merge(layer.All)
	return c
}

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsMutationScopeFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_project": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

// AdminAPIRateLimitsMutationScopeFeatureConfig holds the bucket a mutation
// scope is bounded by. Only per project (app_id): the Admin API authenticates
// as the project, so app_id is the caller identity and the one dimension the
// caller cannot choose. A per-IP bucket would be a weaker proxy for something
// already measured directly -- unlike DCR registration or the CIMD fetch,
// whose endpoints are unauthenticated and where IP is the only handle on the
// caller.
type AdminAPIRateLimitsMutationScopeFeatureConfig struct {
	PerProject *RateLimitConfig `json:"per_project,omitempty"`
}

// SetDefaults mirrors OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig's
// pattern: PerProject is already non-nil by the time this runs
// (SetFieldDefaults force-allocates it), so checking Enabled == nil safely
// detects "no layer configured this bucket" and replaces the whole zero-valued
// struct with the built-in default.
//
// The default is deliberately loose. It is a backstop against runaway or
// abusive volume, not a tuned throttle: no legitimate integration should ever
// have to design around it. Tiers whose projects have no reason to sustain
// scripted Admin API writes are expected to set it far lower.
func (c *AdminAPIRateLimitsMutationScopeFeatureConfig) SetDefaults() {
	if c.PerProject.Enabled == nil {
		c.PerProject = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   1000,
		}
	}
}

// Merge replaces each bucket wholesale, not field-by-field: enabled/period/burst
// are one unit, and merging them field-wise would let two layers jointly
// produce a bucket neither one actually wrote.
func (c *AdminAPIRateLimitsMutationScopeFeatureConfig) Merge(layer *AdminAPIRateLimitsMutationScopeFeatureConfig) *AdminAPIRateLimitsMutationScopeFeatureConfig {
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
	return c
}
