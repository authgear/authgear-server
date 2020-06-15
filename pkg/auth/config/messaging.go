package config

type MessagingConfig struct {
	DefaultEmailMessage EmailMessageConfig `json:"default_email_message,omitempty"`
	DefaultSMSMessage   SMSMessageConfig   `json:"default_sms_message,omitempty"`
	SMSProvider         SMSProvider        `json:"sms_provider,omitempty"`
}

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
)

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
