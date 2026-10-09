package config

type StripeConfig struct {
	// The key starting with "sk_"
	SecretKey string `envconfig:"SECRET_KEY"`

	// The key starting with "whsec_"
	WebhookSigningKey string `envconfig:"WEBHOOK_SIGNING_KEY"`
}

// IsConfigured reports whether this deployment has Stripe at all. Without a
// secret key every Stripe API call fails, so billing features are off.
func (c *StripeConfig) IsConfigured() bool {
	return c != nil && c.SecretKey != ""
}
