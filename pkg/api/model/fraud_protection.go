package model

import "time"

type FraudProtectionDecision string

const (
	FraudProtectionDecisionAllowed FraudProtectionDecision = "allowed"
	FraudProtectionDecisionBlocked FraudProtectionDecision = "blocked"
)

type FraudProtectionAction string

const (
	FraudProtectionActionSendSMS FraudProtectionAction = "send_sms"
)

type FraudProtectionDecisionActionDetail struct {
	Recipient              string `json:"recipient"`
	Type                   string `json:"type"`
	PhoneNumberCountryCode string `json:"phone_number_country_code,omitempty"`
}

type FraudProtectionDecisionRecord struct {
	Timestamp         time.Time                           `json:"timestamp"`
	Decision          FraudProtectionDecision             `json:"decision"`
	Action            FraudProtectionAction               `json:"action"`
	ActionDetail      FraudProtectionDecisionActionDetail `json:"action_detail"`
	TriggeredWarnings []string                            `json:"triggered_warnings"`
	UserAgent         string                              `json:"user_agent,omitempty"`
	IPAddress         string                              `json:"ip_address,omitempty"`
	HTTPUrl           string                              `json:"http_url,omitempty"`
	HTTPReferer       string                              `json:"http_referer,omitempty"`
	UserID            string                              `json:"user_id,omitempty"`
	GeoLocationCode   string                              `json:"geo_location_code,omitempty"`
}
