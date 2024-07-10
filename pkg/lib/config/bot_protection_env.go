package config

type End2EndBotProtectionEnvironmentConfig struct {
	CloudflareEndpoint  string `envconfig:"E2E_BOT_PROTECTION_CLOUDFLARE_ENDPOINT"`
	RecaptchaV2Endpoint string `envconfig:"E2E_BOT_PROTECTION_RECAPTCHAV2_ENDPOINT"`
}
