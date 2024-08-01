package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type FixturePlanName string

const (
	FixtureLimitedPlanName   FixturePlanName = "limited"
	FixtureUnlimitedPlanName FixturePlanName = "unlimited"
)

func FixtureAppConfig(appID string) *config.AppConfig {
	cfg := config.GenerateAppConfigFromOptions(&config.GenerateAppConfigOptions{
		AppID:        appID,
		PublicOrigin: fmt.Sprintf("http://%s.localhost", appID),
	})
	return cfg
}

func FixtureSecretConfig(seed int64) *config.SecretConfig {
	return config.GenerateSecretConfigFromOptions(&config.GenerateSecretConfigOptions{
		DatabaseURL:      "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
		DatabaseSchema:   "public",
		ElasticsearchURL: "http://127.0.0.1:9200",
		RedisURL:         "redis://127.0.0.1",
	}, time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC), rand.New(rand.NewSource(seed)))
}

func FixtureUpdateSecretConfigUpdateInstruction() *config.SecretConfigUpdateInstruction {
	return &config.SecretConfigUpdateInstruction{
		SMTPServerCredentialsUpdateInstruction: &config.SMTPServerCredentialsUpdateInstruction{
			Action: "set",
			Data: &config.SMTPServerCredentialsUpdateInstructionData{
				Host:     "127.0.0.1",
				Port:     25,
				Username: "username",
				Password: "password",
			},
		},
		BotProtectionProviderCredentialsUpdateInstruction: &config.BotProtectionProviderCredentialsUpdateInstruction{
			Action: "set",
			Data: &config.BotProtectionProviderCredentialsUpdateInstructionData{
				Type:      string(config.BotProtectionProviderTypeRecaptchaV2),
				SecretKey: "secret-key",
			},
		},
	}
}

func FixtureFeatureConfig(plan FixturePlanName) *config.FeatureConfig {
	switch plan {
	case FixtureLimitedPlanName:
		fixture := `
oauth:
  client:
    maximum: 1
identity:
  oauth:
    maximum_providers: 1
  biometric:
    disabled: true
hook:
  blocking_handler:
    maximum: 1
  non_blocking_handler:
    maximum: 1
authenticator:
  password:
    policy:
      minimum_guessable_level:
        disabled: true
      excluded_keywords:
        disabled: true
      history:
        disabled: true
`
		cfg, _ := config.ParseFeatureConfig([]byte(fixture))
		return cfg
	case FixtureUnlimitedPlanName:
		fixture := `
oauth:
  client:
    custom_ui_enabled: true
`
		cfg, _ := config.ParseFeatureConfig([]byte(fixture))
		return cfg
	}
	return nil
}
