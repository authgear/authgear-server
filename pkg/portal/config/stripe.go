package config

type StripeConfig struct {
	// The key starting with "sk_"
	SecretKey string `envconfig:"SECRET_KEY"`
}
