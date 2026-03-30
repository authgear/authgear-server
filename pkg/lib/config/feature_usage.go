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
