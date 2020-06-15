package config

import "encoding/json"

type SecretConfig struct {
	Secrets []SecretItem `json:"secrets,omitempty"`
}

type SecretItem struct {
	Key  string          `json:"key,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

type DatabaseCredentials struct {
	DatabaseURL    string `json:"database_url,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
}

type SMTPMode string

const (
	SMTPModeNormal SMTPMode = "normal"
	SMTPModeSSL    SMTPMode = "ssl"
)

type SMTPServerCredentials struct {
	Host     string   `json:"host,omitempty"`
	Port     int      `json:"port,omitempty"`
	Mode     SMTPMode `json:"mode,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}

type TwilioCredentials struct {
	AccountSID string `json:"account_sid,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
}

type NexmoCredentials struct {
	APIKey    string `json:"api_key,omitempty"`
	APISecret string `json:"api_secret,omitempty"`
}
