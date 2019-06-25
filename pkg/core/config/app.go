package config

// AppConfiguration is configuration kept secret from the developer.
type AppConfiguration struct {
	Version     string                 `json:"version" yaml:"version" msg:"version"`
	DatabaseURL string                 `json:"database_url" yaml:"database_url" msg:"database_url"`
	SMTP        NewSMTPConfiguration   `json:"smtp" yaml:"smtp" msg:"smtp"`
	Twilio      NewTwilioConfiguration `json:"twilio" yaml:"twilio" msg:"twilio"`
	Nexmo       NewNexmoConfiguration  `json:"nexmo" yaml:"nexmo" msg:"nexmo"`
}

type NewSMTPConfiguration struct {
	Host     string `json:"host" yaml:"host" msg:"host"`
	Port     int    `json:"port" yaml:"port" msg:"port"`
	Mode     string `json:"mode" yaml:"mode" msg:"mode"`
	Login    string `json:"login" yaml:"login" msg:"login"`
	Password string `json:"password" yaml:"password" msg:"password"`
}

type NewTwilioConfiguration struct {
	AccountSID string `json:"account_sid" yaml:"account_sid" msg:"account_sid"`
	AuthToken  string `json:"auth_token" yaml:"auth_token" msg:"auth_token"`
	From       string `json:"from" yaml:"from" msg:"from"`
}

type NewNexmoConfiguration struct {
	APIKey    string `json:"api_key" yaml:"api_key" msg:"api_key"`
	APISecret string `json:"secret" yaml:"secret" msg:"secret"`
	From      string `json:"from" yaml:"from" msg:"from"`
}
