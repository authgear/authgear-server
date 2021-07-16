package test

import (
	"fmt"
	"math/rand"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type FixturePlanName string

const (
	FixtureLimitedPlanName   FixturePlanName = "limited"
	FixtureUnlimitedPlanName FixturePlanName = "unlimited"
)

func newInt(v int) *int { return &v }

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
	}, rand.New(rand.NewSource(seed)))
}

func FixtureFeatureConfig(plan FixturePlanName) *config.FeatureConfig {
	switch plan {
	case FixtureLimitedPlanName:
		return &config.FeatureConfig{
			OAuth: &config.OAuthFeatureConfig{
				Client: &config.OAuthClientFeatureConfig{
					Maximum: newInt(1),
				},
			},
		}
	case FixtureUnlimitedPlanName:
		return config.NewEffectiveDefaultFeatureConfig()
	}
	return nil
}
