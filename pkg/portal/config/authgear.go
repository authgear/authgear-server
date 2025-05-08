package config

type AuthgearConfig struct {
	ClientID string `envconfig:"CLIENT_ID"`
	Endpoint string `envconfig:"ENDPOINT"`
	AppID    string `envconfig:"APP_ID"`

	WebSDKSessionType string `envconfig:"WEB_SDK_SESSION_TYPE" default:"cookie"`

	OnceLicenseKey      string `envconfig:"ONCE_LICENSE_KEY"`
	OnceLicenseExpireAt string `envconfig:"ONCE_LICENSE_EXPIRE_AT"`
	OnceLicenseeEmail   string `envconfig:"ONCE_LICENSEE_EMAIL"`
}
