package config

import "github.com/authgear/authgear-server/pkg/api/model"

var _ = FeatureConfigSchema.Add("UsageMatch", `
{
	"type": "string",
	"enum": ["*", "user_export", "user_import", "email", "whatsapp", "sms", "oauth_client_dcr", "oauth_client_cimd"]
}
`)

var _ = FeatureConfigSchema.Add("UsageLimitAction", `
{
	"type": "string",
	"enum": ["alert", "block"]
}
`)

var _ = FeatureConfigSchema.Add("FeatureUsageLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"quota": { "type": "integer", "minimum": 0 },
		"period": { "$ref": "#/$defs/UsageLimitPeriod" },
		"action": { "$ref": "#/$defs/UsageLimitAction" }
	},
	"required": ["quota", "period", "action"]
}
`)

type FeatureUsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Period model.UsageLimitPeriod `json:"period"`
	Action model.UsageLimitAction `json:"action"`
}

// StandingFeatureUsageLimitConfig is a separate shape from
// FeatureUsageLimitConfig because a standing limit has no period: its
// "usage" is a live count (e.g. COUNT(*) of DCR clients), not a
// periodically-reset counter. See pkg/lib/usage/standing.go.
var _ = FeatureConfigSchema.Add("StandingFeatureUsageLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"quota": { "type": "integer", "minimum": 0 },
		"action": { "$ref": "#/$defs/UsageLimitAction" }
	},
	"required": ["quota", "action"]
}
`)

type StandingFeatureUsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Action model.UsageLimitAction `json:"action"`
}

var _ = FeatureConfigSchema.Add("FeatureUsageLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"user_export": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"user_import": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"email": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"whatsapp": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"sms": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"oauth_client_dcr": { "type": "array", "items": { "$ref": "#/$defs/StandingFeatureUsageLimitConfig" } }
	}
}
`)

type FeatureUsageLimitsConfig struct {
	// omitzero, not omitempty: nil (unset, inherit from a lower layer) and
	// an explicit empty slice (override to "no limits") are semantically
	// distinct here — omitempty would drop both cases the same way,
	// silently reverting an explicit override back to inherited on the
	// merge-fold's marshal/re-parse round trip (see
	// viewEffectiveResource in configsource/resources.go). omitzero only
	// omits the true zero value (nil), preserving an explicit []
	// through that round trip.
	UserExport     []FeatureUsageLimitConfig         `json:"user_export,omitzero"`
	UserImport     []FeatureUsageLimitConfig         `json:"user_import,omitzero"`
	Email          []FeatureUsageLimitConfig         `json:"email,omitzero"`
	Whatsapp       []FeatureUsageLimitConfig         `json:"whatsapp,omitzero"`
	SMS            []FeatureUsageLimitConfig         `json:"sms,omitzero"`
	OAuthClientDCR []StandingFeatureUsageLimitConfig `json:"oauth_client_dcr,omitzero"`
}

func (c *FeatureUsageLimitsConfig) Limits(name model.UsageName) []FeatureUsageLimitConfig {
	if c == nil {
		return nil
	}

	switch name {
	case model.UsageNameUserExport:
		return c.UserExport
	case model.UsageNameUserImport:
		return c.UserImport
	case model.UsageNameEmail:
		return c.Email
	case model.UsageNameWhatsapp:
		return c.Whatsapp
	case model.UsageNameSMS:
		return c.SMS
	default:
		return nil
	}
}

// StandingLimits is the parallel accessor for standing (period-less) usage
// names, since the element type (StandingFeatureUsageLimitConfig) differs
// from Limits' FeatureUsageLimitConfig.
func (c *FeatureUsageLimitsConfig) StandingLimits(name model.UsageName) []StandingFeatureUsageLimitConfig {
	if c == nil {
		return nil
	}

	switch name {
	case model.UsageNameOAuthClientDCR:
		return c.OAuthClientDCR
	default:
		return nil
	}
}

var _ = FeatureConfigSchema.Add("FeatureUsageHookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string", "format": "x_hook_uri" },
		"match": { "$ref": "#/$defs/UsageMatch" }
	},
	"required": ["url", "match"]
}
`)

type FeatureUsageHookConfig struct {
	URL   string `json:"url"`
	Match string `json:"match"`
}

var _ = FeatureConfigSchema.Add("FeatureUsageConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"hooks": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageHookConfig" } },
		"limits": { "$ref": "#/$defs/FeatureUsageLimitsConfig" }
	}
}
`)

type FeatureUsageConfig struct {
	Hooks  []FeatureUsageHookConfig  `json:"hooks,omitempty"`
	Limits *FeatureUsageLimitsConfig `json:"limits,omitempty" nullable:"true"`
}

var _ MergeableFeatureConfig = &FeatureUsageConfig{}

func (c *FeatureUsageConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Usage == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &FeatureUsageConfig{}
	}

	if len(layer.Usage.Hooks) > 0 {
		merged.Hooks = append(append([]FeatureUsageHookConfig{}, merged.Hooks...), layer.Usage.Hooks...)
	}

	if layer.Usage.Limits != nil {
		merged.Limits = mergeFeatureUsageLimits(merged.Limits, layer.Usage.Limits)
	}

	return merged
}

func (c *FeatureConfig) migrateDeprecatedUsageLimit(name model.UsageName, legacy *Deprecated_UsageLimitConfig) {
	if legacy == nil || !legacy.IsEnabled() {
		return
	}
	if c.featureUsageLimits(name) != nil {
		return
	}

	limit := FeatureUsageLimitConfig{
		Quota:  legacy.GetQuota(),
		Period: model.UsageLimitPeriod(legacy.Period),
		Action: model.UsageLimitActionBlock,
	}

	c.ensureFeatureUsageLimits()
	c.setFeatureUsageLimits(name, []FeatureUsageLimitConfig{limit})
}

func (c *FeatureConfig) ensureFeatureUsageLimits() {
	if c.Usage == nil {
		c.Usage = &FeatureUsageConfig{}
	}
	if c.Usage.Limits == nil {
		c.Usage.Limits = &FeatureUsageLimitsConfig{}
	}
}

func (c *FeatureConfig) featureUsageLimits(name model.UsageName) []FeatureUsageLimitConfig {
	if c == nil || c.Usage == nil || c.Usage.Limits == nil {
		return nil
	}
	return c.Usage.Limits.Limits(name)
}

func (c *FeatureConfig) setFeatureUsageLimits(name model.UsageName, limits []FeatureUsageLimitConfig) {
	switch name {
	case model.UsageNameUserExport:
		c.Usage.Limits.UserExport = limits
	case model.UsageNameUserImport:
		c.Usage.Limits.UserImport = limits
	case model.UsageNameEmail:
		c.Usage.Limits.Email = limits
	case model.UsageNameWhatsapp:
		c.Usage.Limits.Whatsapp = limits
	case model.UsageNameSMS:
		c.Usage.Limits.SMS = limits
	}
}

func mergeFeatureUsageLimits(base *FeatureUsageLimitsConfig, layer *FeatureUsageLimitsConfig) *FeatureUsageLimitsConfig {
	if layer == nil {
		return base
	}

	merged := base
	if merged == nil {
		merged = &FeatureUsageLimitsConfig{}
	}

	if layer.UserExport != nil {
		merged.UserExport = layer.UserExport
	}
	if layer.UserImport != nil {
		merged.UserImport = layer.UserImport
	}
	if layer.Email != nil {
		merged.Email = layer.Email
	}
	if layer.Whatsapp != nil {
		merged.Whatsapp = layer.Whatsapp
	}
	if layer.SMS != nil {
		merged.SMS = layer.SMS
	}
	if layer.OAuthClientDCR != nil {
		merged.OAuthClientDCR = layer.OAuthClientDCR
	}

	return merged
}
