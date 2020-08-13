package config

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func NewAppConfigFromOptions(opts *Options) *config.AppConfig {
	return &config.AppConfig{
		ID:   config.AppID(opts.AppID),
		HTTP: &config.HTTPConfig{Hosts: opts.Hosts},
	}
}

func NewSecretConfigFromOptions(opts *SecretOptions) *config.SecretConfig {
	var items []config.SecretItem

	items = append(items, config.SecretItem{
		Key: config.DatabaseCredentialsKey,
		Data: &config.DatabaseCredentials{
			DatabaseURL:    opts.DatabaseURL,
			DatabaseSchema: opts.DatabaseSchema,
		},
	})
	items = append(items, config.SecretItem{
		Key: config.RedisCredentialsKey,
		Data: &config.RedisCredentials{
			Host:     opts.RedisHost,
			Port:     opts.RedisPort,
			Password: opts.RedisPassword,
		},
	})

	items = append(items, config.SecretItem{
		Key:  config.OIDCKeyMaterialsKey,
		Data: &config.OIDCKeyMaterials{Set: generateRSAKey()},
	})

	items = append(items, config.SecretItem{
		Key:  config.CSRFKeyMaterialsKey,
		Data: &config.CSRFKeyMaterials{Set: generateOctetKey()},
	})

	items = append(items, config.SecretItem{
		Key:  config.WebhookKeyMaterialsKey,
		Data: &config.WebhookKeyMaterials{Set: generateOctetKey()},
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

func generateOctetKey() jwk.Set {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(key)
	if err != nil {
		panic(err)
	}

	jwkKey.Set(jwk.KeyIDKey, uuid.New())

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	return keySet
}

func generateRSAKey() jwk.Set {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(privateKey)
	if err != nil {
		panic(err)
	}
	jwkKey.Set(jwk.KeyIDKey, uuid.New())
	jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	jwkKey.Set(jwk.AlgorithmKey, "RS256")

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	return keySet
}
