package test

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func FixtureAppConfig(appID string) *config.AppConfig {
	appConfig, err := config.Parse([]byte(fmt.Sprintf(`id: %q`, appID)))
	if err != nil {
		panic(err)
	}
	return appConfig
}

func FixtureSecretConfig(seed int64) *config.SecretConfig {
	var items []config.SecretItem
	random := rand.New(rand.NewSource(seed))

	items = append(items, config.SecretItem{
		Key: config.DatabaseCredentialsKey,
		Data: &config.DatabaseCredentials{
			DatabaseURL:    "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
			DatabaseSchema: "public",
		},
	})
	items = append(items, config.SecretItem{
		Key: config.RedisCredentialsKey,
		Data: &config.RedisCredentials{
			RedisURL: "redis://127.0.0.1",
		},
	})

	items = append(items, config.SecretItem{
		Key:  config.OIDCKeyMaterialsKey,
		Data: &config.OIDCKeyMaterials{Set: generateRSAKey(random)},
	})

	items = append(items, config.SecretItem{
		Key:  config.CSRFKeyMaterialsKey,
		Data: &config.CSRFKeyMaterials{Set: generateOctetKey(random)},
	})

	items = append(items, config.SecretItem{
		Key:  config.AdminAPIAuthKeyKey,
		Data: &config.AdminAPIAuthKey{Set: generateRSAKey(random)},
	})

	marshalSecretData(items)
	return &config.SecretConfig{Secrets: items}
}

func marshalSecretData(items []config.SecretItem) {
	for i, item := range items {
		data, err := json.Marshal(item.Data)
		if err != nil {
			panic(err)
		}

		item.RawData = data
		items[i] = item
	}
}

func generateOctetKey(rand io.Reader) jwk.Set {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(key)
	if err != nil {
		panic(err)
	}

	_ = jwkKey.Set(jwk.KeyIDKey, "key-id")

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	return keySet
}

func generateRSAKey(rand io.Reader) jwk.Set {
	privateKey, err := rsa.GenerateKey(rand, 2048)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(privateKey)
	if err != nil {
		panic(err)
	}
	_ = jwkKey.Set(jwk.KeyIDKey, "key-id")
	_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	return keySet
}
