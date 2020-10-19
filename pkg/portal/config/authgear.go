package config

type AuthgearConfig struct {
	ClientID string `envconfig:"CLIENT_ID"`
	Endpoint string `envconfig:"ENDPOINT"`
	AppID    string `envconfig:"APP_ID"`
}
