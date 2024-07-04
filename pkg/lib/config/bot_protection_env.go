package config

type BotProtectionEnvironmentConfig struct {
	CloudflareEndpoint  string `envconfig:"BOT_PROTECTION_CLOUDFLARE_ENDPOINT" default:"https://challenges.cloudflare.com/turnstile/v0/siteverify"`
	RecaptchaV2Endpoint string `envconfig:"BOT_PROTECTION_RECAPTCHAV2_ENDPOINT" default:"https://www.google.com/recaptcha/api/siteverify"`
}

// NewDefaultBotProtectionEnvironmentConfig provides default bot protection config
func NewDefaultBotProtectionEnvironmentConfig() *BotProtectionEnvironmentConfig {
	return &BotProtectionEnvironmentConfig{
		CloudflareEndpoint:  "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		RecaptchaV2Endpoint: "https://www.google.com/recaptcha/api/siteverify",
	}
}
