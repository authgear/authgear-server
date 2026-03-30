package config

import "github.com/authgear/authgear-server/pkg/api/model"

type UsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Period model.UsageLimitPeriod `json:"period"`
	Action model.UsageLimitAction `json:"action"`
}

type UsageLimitsConfig struct {
	UserExport []UsageLimitConfig `json:"user_export,omitempty"`
	UserImport []UsageLimitConfig `json:"user_import,omitempty"`
	Email      []UsageLimitConfig `json:"email,omitempty"`
	Whatsapp   []UsageLimitConfig `json:"whatsapp,omitempty"`
	SMS        []UsageLimitConfig `json:"sms,omitempty"`
}

func (c *UsageLimitsConfig) Limits(name model.UsageName) []UsageLimitConfig {
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

type UsageAlertConfig struct {
	Type  string `json:"type"`
	Email string `json:"email,omitempty"`
	Match string `json:"match"`
}

type UsageConfig struct {
	Alerts []UsageAlertConfig `json:"alerts,omitempty"`
	Limits *UsageLimitsConfig `json:"limits,omitempty" nullable:"true"`
}
