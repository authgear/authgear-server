package test

import (
	"fmt"
	"math/rand"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
		DatabaseURL:    "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
		DatabaseSchema: "public",
		RedisURL:       "redis://127.0.0.1",
	}, rand.New(rand.NewSource(seed)))
}
