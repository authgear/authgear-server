package config

// AppConfiguration is configuration kept secret from the developer.
type AppConfiguration struct {
	Version     string                 `json:"version"`
	DatabaseURL string                 `json:"database_url"`
	SMTP        NewSMTPConfiguration   `json:"smtp"`
	Twilio      NewTwilioConfiguration `json:"twilio"`
	Nexmo       NewNexmoConfiguration  `json:"nexmo"`
}

type NewSMTPConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Mode     string `json:"mode"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type NewTwilioConfiguration struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	From       string `json:"from"`
}

type NewNexmoConfiguration struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"secret"`
	From      string `json:"from"`
}
