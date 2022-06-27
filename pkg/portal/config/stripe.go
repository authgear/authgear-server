package config

type StripeConfig struct {
	// The key starting with "sk_"
	SecretKey string `envconfig:"SECRET_KEY"`

	// The key starting with "whsec_"
	WebhookSigningKey string `envconfig:"WEBHOOK_SIGNING_KEY"`
}
