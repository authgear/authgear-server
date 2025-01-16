package model

type SMSProviderConfigurationInput struct {
	Twilio  *SMSProviderConfigurationTwilioInput  `json:"twilio,omitempty"`
	Webhook *SMSProviderConfigurationWebhookInput `json:"webhook,omitempty"`
}

type SMSProviderConfigurationTwilioInput struct {
	AccountSID          string `json:"accountSID,omitempty"`
	AuthToken           string `json:"authToken,omitempty"`
	MessagingServiceSID string `json:"messagingServiceSID,omitempty"`
}

type SMSProviderConfigurationWebhookInput struct {
	URL     string `json:"url,omitempty"`
	Timeout *int   `json:"timeout,omitempty"`
}
