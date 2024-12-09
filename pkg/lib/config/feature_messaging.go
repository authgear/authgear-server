package config

var _ = FeatureConfigSchema.Add("MessagingFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"rate_limits": { "$ref": "#/$defs/MessagingRateLimitsFeatureConfig" },
		"sms_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"email_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"whatsapp_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"sms_usage_count_disabled": { "type": "boolean" },
		"whatsapp_usage_count_disabled": { "type": "boolean" },
		"template_customization_disabled": { "type": "boolean" },
		"custom_smtp_disabled": { "type": "boolean" }
	}
}
`)

type MessagingFeatureConfig struct {
	RateLimits *MessagingRateLimitsFeatureConfig `json:"rate_limits,omitempty"`

	SMSUsage      *UsageLimitConfig `json:"sms_usage,omitempty"`
	EmailUsage    *UsageLimitConfig `json:"email_usage,omitempty"`
	WhatsappUsage *UsageLimitConfig `json:"whatsapp_usage,omitempty"`

	SMSUsageCountDisabled      *bool `json:"sms_usage_count_disabled,omitempty"`
	WhatsappUsageCountDisabled *bool `json:"whatsapp_usage_count_disabled,omitempty"`

	TemplateCustomizationDisabled *bool `json:"template_customization_disabled,omitempty"`

	CustomSMTPDisabled *bool `json:"custom_smtp_disabled,omitempty"`
}

var _ MergeableFeatureConfig = &MessagingFeatureConfig{}

func (c *MessagingFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Messaging == nil {
		return c
	}
	var merged *MessagingFeatureConfig = c
	if merged == nil {
		merged = &MessagingFeatureConfig{}
	}
	if layer.Messaging.RateLimits != nil {
		merged.RateLimits = layer.Messaging.RateLimits
	}
	if layer.Messaging.SMSUsage != nil {
		merged.SMSUsage = layer.Messaging.SMSUsage
	}
	if layer.Messaging.EmailUsage != nil {
		merged.EmailUsage = layer.Messaging.EmailUsage
	}
	if layer.Messaging.WhatsappUsage != nil {
		merged.WhatsappUsage = layer.Messaging.WhatsappUsage
	}
	if layer.Messaging.SMSUsageCountDisabled != nil {
		merged.SMSUsageCountDisabled = layer.Messaging.SMSUsageCountDisabled
	}
	if layer.Messaging.WhatsappUsageCountDisabled != nil {
		merged.WhatsappUsageCountDisabled = layer.Messaging.WhatsappUsageCountDisabled
	}
	if layer.Messaging.TemplateCustomizationDisabled != nil {
		merged.TemplateCustomizationDisabled = layer.Messaging.TemplateCustomizationDisabled
	}
	if layer.Messaging.CustomSMTPDisabled != nil {
		merged.CustomSMTPDisabled = layer.Messaging.CustomSMTPDisabled
	}

	return merged
}

func (c *MessagingFeatureConfig) SetDefaults() {
	if c.SMSUsage.Enabled == nil {
		c.SMSUsage = &UsageLimitConfig{
			Enabled: newBool(false),
		}
	}
	if c.EmailUsage.Enabled == nil {
		c.EmailUsage = &UsageLimitConfig{
			Enabled: newBool(false),
		}
	}
	if c.WhatsappUsage.Enabled == nil {
		c.WhatsappUsage = &UsageLimitConfig{
			Enabled: newBool(false),
		}
	}

	if c.SMSUsageCountDisabled == nil {
		c.SMSUsageCountDisabled = newBool(false)
	}
	if c.WhatsappUsageCountDisabled == nil {
		c.WhatsappUsageCountDisabled = newBool(false)
	}
	if c.TemplateCustomizationDisabled == nil {
		c.TemplateCustomizationDisabled = newBool(false)
	}
	if c.CustomSMTPDisabled == nil {
		c.CustomSMTPDisabled = newBool(false)
	}
}

var _ = FeatureConfigSchema.Add("MessagingRateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_target": { "$ref": "#/$defs/RateLimitConfig" },
		"email": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_target": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type MessagingRateLimitsFeatureConfig struct {
	SMS            *RateLimitConfig `json:"sms,omitempty"`
	SMSPerIP       *RateLimitConfig `json:"sms_per_ip,omitempty"`
	SMSPerTarget   *RateLimitConfig `json:"sms_per_target,omitempty"`
	Email          *RateLimitConfig `json:"email,omitempty"`
	EmailPerIP     *RateLimitConfig `json:"email_per_ip,omitempty"`
	EmailPerTarget *RateLimitConfig `json:"email_per_target,omitempty"`
}

func (c *MessagingRateLimitsFeatureConfig) SetDefaults() {
	if c.SMSPerIP.Enabled == nil {
		c.SMSPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
	if c.SMSPerTarget.Enabled == nil {
		c.SMSPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1h",
			Burst:   10,
		}
	}
	if c.EmailPerIP.Enabled == nil {
		c.EmailPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   200,
		}
	}
	if c.EmailPerTarget.Enabled == nil {
		c.EmailPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "24h",
			Burst:   50,
		}
	}
}
