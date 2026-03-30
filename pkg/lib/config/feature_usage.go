package config

import "github.com/authgear/authgear-server/pkg/api/model"

type FeatureUsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Period model.UsageLimitPeriod `json:"period"`
	Action model.UsageLimitAction `json:"action"`
}

type FeatureUsageLimitsConfig struct {
	UserExport []FeatureUsageLimitConfig `json:"user_export,omitempty"`
	UserImport []FeatureUsageLimitConfig `json:"user_import,omitempty"`
	Email      []FeatureUsageLimitConfig `json:"email,omitempty"`
	Whatsapp   []FeatureUsageLimitConfig `json:"whatsapp,omitempty"`
	SMS        []FeatureUsageLimitConfig `json:"sms,omitempty"`
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

type FeatureUsageHookConfig struct {
	URL   string `json:"url"`
	Match string `json:"match"`
}

type FeatureUsageConfig struct {
	Hooks  []FeatureUsageHookConfig  `json:"hooks,omitempty"`
	Limits *FeatureUsageLimitsConfig `json:"limits,omitempty" nullable:"true"`
}

func (c *FeatureConfig) Migrate() *FeatureConfig {
	if c == nil {
		return nil
	}

	if c.Messaging != nil {
		c.migrateDeprecatedUsageLimit(model.UsageNameSMS, c.Messaging.SMSUsage)
		c.migrateDeprecatedUsageLimit(model.UsageNameEmail, c.Messaging.EmailUsage)
		c.migrateDeprecatedUsageLimit(model.UsageNameWhatsapp, c.Messaging.WhatsappUsage)
	}
	if c.AdminAPI != nil {
		c.migrateDeprecatedUsageLimit(model.UsageNameUserImport, c.AdminAPI.UserImportUsage)
		c.migrateDeprecatedUsageLimit(model.UsageNameUserExport, c.AdminAPI.UserExportUsage)
	}

	return c
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
