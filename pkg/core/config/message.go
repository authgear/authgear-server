package config

//go:generate msgp -tests=false

type MessagesConfiguration struct {
	Email       EmailMessageConfiguration `json:"email,omitempty" yaml:"email" msg:"email"`
	SMSProvider SMSProvider               `json:"sms_provider,omitempty" yaml:"sms_provider,omitempty" msg:"sms_provider"`
	SMS         SMSMessageConfiguration   `json:"sms,omitempty" yaml:"sms" msg:"sms"`
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

func (c EmailMessageConfiguration) Sender() string {
	return c["sender"]
}

func (c EmailMessageConfiguration) Subject() string {
	return c["subject"]
}

func (c EmailMessageConfiguration) ReplyTo() string {
	return c["reply_to"]
}

func (c EmailMessageConfiguration) SetSender(sender string) {
	c["sender"] = sender
}

func (c EmailMessageConfiguration) SetSubject(subject string) {
	c["subject"] = subject
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

func (c SMSMessageConfiguration) Sender() string {
	return c["sender"]
}
