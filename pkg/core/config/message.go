package config

//go:generate msgp -tests=false

type MessagesConfiguration struct {
	Email       EmailMessageConfiguration `json:"email,omitempty" yaml:"email" msg:"email" default_zero_value:"true"`
	SMSProvider SMSProvider               `json:"sms_provider,omitempty" yaml:"sms_provider,omitempty" msg:"sms_provider"`
	SMS         SMSMessageConfiguration   `json:"sms,omitempty" yaml:"sms" msg:"sms" default_zero_value:"true"`
}

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
)

type EmailMessageConfiguration map[string]string

func NewEmailMessageConfiguration(configs ...EmailMessageConfiguration) EmailMessageConfiguration {
	config := EmailMessageConfiguration{}
	for _, c := range configs {
		for k, v := range c {
			config[k] = v
		}
	}
	return config
}

type SMSMessageConfiguration map[string]string

func NewSMSMessageConfiguration(configs ...SMSMessageConfiguration) SMSMessageConfiguration {
	config := SMSMessageConfiguration{}
	for _, c := range configs {
		for k, v := range c {
			config[k] = v
		}
	}
	return config
}
