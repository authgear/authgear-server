package model

type SMSProviderConfigurationInput struct {
	Twilio *SMSProviderConfigurationTwilioInput `json:"twilio,omitempty"`
}

type SMSProviderConfigurationTwilioInput struct {
	AccountSID          string `json:"accountSID,omitempty"`
	AuthToken           string `json:"authToken,omitempty"`
	MessagingServiceSID string `json:"messagingServiceSID,omitempty"`
}
