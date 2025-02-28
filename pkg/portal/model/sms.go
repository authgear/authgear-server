package model

import "github.com/authgear/authgear-server/pkg/lib/config"

type SMSProviderConfigurationInput struct {
	Twilio  *SMSProviderConfigurationTwilioInput  `json:"twilio,omitempty"`
	Webhook *SMSProviderConfigurationWebhookInput `json:"webhook,omitempty"`
	Deno    *SMSProviderConfigurationDenoInput    `json:"deno,omitempty"`
}

type SMSProviderConfigurationTwilioInput struct {
	CredentialType      config.TwilioCredentialType `json:"credentialType,omitempty"`
	AccountSID          string                      `json:"accountSID,omitempty"`
	AuthToken           string                      `json:"authToken,omitempty"`
	APIKeySID           string                      `json:"apiKeySID,omitempty"`
	APIKeySecret        string                      `json:"apiKeySecret,omitempty"`
	MessagingServiceSID *string                     `json:"messagingServiceSID,omitempty"`
}

type SMSProviderConfigurationWebhookInput struct {
	URL     string `json:"url,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}

type SMSProviderConfigurationDenoInput struct {
	Script  string `json:"script,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}
