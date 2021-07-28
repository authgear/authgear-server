package config

import (
	"encoding/json"
	mathrand "math/rand"
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/util/secrets"
)

type GenerateAppConfigOptions struct {
	AppID        string
	PublicOrigin string
	CookieDomain string
}

func GenerateAppConfigFromOptions(opts *GenerateAppConfigOptions) *AppConfig {
	cfg := &AppConfig{
		ID:   AppID(opts.AppID),
		HTTP: &HTTPConfig{PublicOrigin: opts.PublicOrigin},
	}
	if opts.CookieDomain != "" {
		cfg.HTTP.CookieDomain = &opts.CookieDomain
	}
	return cfg
}

type GenerateSecretConfigOptions struct {
	DatabaseURL      string
	DatabaseSchema   string
	ElasticsearchURL string
	RedisURL         string
}

func GenerateSecretConfigFromOptions(opts *GenerateSecretConfigOptions, createdAt time.Time, rng *mathrand.Rand) *SecretConfig {
	var items []SecretItem

	if opts.DatabaseURL != "" {
		items = append(items, SecretItem{
			Key: DatabaseCredentialsKey,
			Data: &DatabaseCredentials{
				DatabaseURL:    opts.DatabaseURL,
				DatabaseSchema: opts.DatabaseSchema,
			},
		})
	}
	if opts.ElasticsearchURL != "" {
		items = append(items, SecretItem{
			Key: ElasticsearchCredentialsKey,
			Data: &ElasticsearchCredentials{
				ElasticsearchURL: opts.ElasticsearchURL,
			},
		})
	}
	if opts.RedisURL != "" {
		items = append(items, SecretItem{
			Key: RedisCredentialsKey,
			Data: &RedisCredentials{
				RedisURL: opts.RedisURL,
			},
		})
	}

	items = append(items, SecretItem{
		Key:  OAuthKeyMaterialsKey,
		Data: &OAuthKeyMaterials{Set: wrapInSet(secrets.GenerateRSAKey(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  CSRFKeyMaterialsKey,
		Data: &CSRFKeyMaterials{Set: wrapInSet(secrets.GenerateOctetKey(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  WebhookKeyMaterialsKey,
		Data: &WebhookKeyMaterials{Set: wrapInSet(secrets.GenerateOctetKey(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  AdminAPIAuthKeyKey,
		Data: &AdminAPIAuthKey{Set: wrapInSet(secrets.GenerateRSAKey(createdAt, rng))},
	})

	marshalSecretData(items)
	return &SecretConfig{Secrets: items}
}

func marshalSecretData(items []SecretItem) {
	for i, item := range items {
		data, err := json.Marshal(item.Data)
		if err != nil {
			panic(err)
		}

		item.RawData = data
		items[i] = item
	}
}

func wrapInSet(jwkKey jwk.Key) jwk.Set {
	keySet := jwk.NewSet()
	_ = keySet.Add(jwkKey)
	return keySet
}
