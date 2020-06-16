package config

var _ = Schema.Add("MessagingConfig", `
{
	"type": "object",
	"properties": {
		"default_email_message": { "$ref": "#/$defs/EmailMessageConfig" },
		"default_sms_message": { "$ref": "#/$defs/SMSMessageConfig" },
		"sms_provider": { "$ref": "#/$defs/SMSProvider" }
	}
}
`)

type MessagingConfig struct {
	DefaultEmailMessage EmailMessageConfig `json:"default_email_message,omitempty"`
	DefaultSMSMessage   SMSMessageConfig   `json:"default_sms_message,omitempty"`
	SMSProvider         SMSProvider        `json:"sms_provider,omitempty"`
}

func (c *MessagingConfig) SetDefaults() {
	if c.DefaultEmailMessage["sender"] == "" {
		c.DefaultEmailMessage["sender"] = "no-reply@skygear.io"
	}
}

var _ = Schema.Add("SMSProvider", `
{
	"type": "string",
	"enum": ["nexmo", "twilio"]
}
`)

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
)

var _ = Schema.Add("EmailMessageConfig", `
{
	"type": "object",
	"properties": {
		"^sender(#.+)?$": { "type": "string", "format": "idn-email" },
		"^subject(#.+)?$": { "type": "string" },
		"^reply_to(#.+)?$": { "type": "string", "format": "idn-email" }
	}
}
`)

type EmailMessageConfig map[string]string

func NewEmailMessageConfig(configs ...EmailMessageConfig) EmailMessageConfig {
	config := EmailMessageConfig{}
	for _, c := range configs {
		for k, v := range c {
			config[k] = v
		}
	}
	return config
}

var _ = Schema.Add("SMSMessageConfig", `
{
	"type": "object",
	"properties": {
		"^sender(#.+)?$": { "type": "string", "format": "phone" }
	}
}
`)

type SMSMessageConfig map[string]string

func NewSMSMessageConfig(configs ...SMSMessageConfig) SMSMessageConfig {
	config := SMSMessageConfig{}
	for _, c := range configs {
		for k, v := range c {
			config[k] = v
		}
	}
	return config
}
