package config

type AppConfig struct {
	HostTemplate string              `envconfig:"HOST_TEMPLATE"`
	Secret       AppSecretConfig     `envconfig:"SECRET"`
	Kubernetes   AppKubernetesConfig `envconfig:"KUBERNETES"`
}

type AppKubernetesConfig struct {
	NewResourcePrefix string `envconfig:"NEW_RESOURCE_PREFIX"`
}

type AppSecretConfig struct {
	DatabaseURL      string `envconfig:"DATABASE_URL"`
	DatabaseSchema   string `envconfig:"DATABASE_SCHEMA"`
	RedisURL         string `envconfig:"REDIS_URL"`
	SMTPHost         string `envconfig:"SMTP_HOST"`
	SMTPPort         int    `envconfig:"SMTP_PORT"`
	SMTPMode         string `envconfig:"SMTP_MODE"`
	SMTPUsername     string `envconfig:"SMTP_USERNAME"`
	SMTPPassword     string `envconfig:"SMTP_PASSWORD"`
	TwilioAccountSID string `envconfig:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `envconfig:"TWILIO_AUTH_TOKEN"`
	NexmoAPIKey      string `envconfig:"NEXMO_API_KEY"`
	NexmoAPISecret   string `envconfig:"NEXMO_API_SECRET"`
}
