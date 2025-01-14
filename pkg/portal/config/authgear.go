package config

type AuthgearConfig struct {
	ClientID string `envconfig:"CLIENT_ID"`
	Endpoint string `envconfig:"ENDPOINT"`
	AppID    string `envconfig:"APP_ID"`

	WebSDKSessionType string `envconfig:"WEB_SDK_SESSION_TYPE" default:"cookie"`
}
