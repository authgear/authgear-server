package config

import (
	"crypto/rsa"
	"encoding/base32"
	"encoding/json"
	"io"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/util/uuid"
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

func GenerateSecretConfigFromOptions(opts *GenerateSecretConfigOptions, rand io.Reader) *SecretConfig {
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
		Data: &OAuthKeyMaterials{Set: generateRSAKey(rand)},
	})

	items = append(items, SecretItem{
		Key:  CSRFKeyMaterialsKey,
		Data: &CSRFKeyMaterials{Set: generateOctetKey(rand)},
	})

	items = append(items, SecretItem{
		Key:  WebhookKeyMaterialsKey,
		Data: &WebhookKeyMaterials{Set: generateOctetKey(rand)},
	})

	items = append(items, SecretItem{
		Key:  AdminAPIAuthKeyKey,
		Data: &AdminAPIAuthKey{Set: generateRSAKey(rand)},
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

func generateKeyID(rand io.Reader) string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id)
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

	id := make([]byte, 16)
	_, err = rand.Read(id)
	if err != nil {
		panic(err)
	}

	_ = jwkKey.Set(jwk.KeyIDKey, generateKeyID(rand))

	keySet := jwk.NewSet()
	_ = keySet.Add(jwkKey)
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
	_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
	_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

	keySet := jwk.NewSet()
	_ = keySet.Add(jwkKey)
	return keySet
}
