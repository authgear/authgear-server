package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	mathrand "math/rand"
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
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

type GenerateOAuthClientConfigOptions struct {
	Name                  string
	ApplicationType       OAuthClientApplicationType
	RedirectURI           string
	PostLogoutRedirectURI string
}

func GenerateOAuthConfigFromOptions(opts *GenerateOAuthClientConfigOptions) (*OAuthClientConfig, error) {
	if opts.ApplicationType == OAuthClientApplicationTypeThirdPartyApp {
		// third-party apps require client secret
		return nil, errors.New("generating third-party apps is not supported")
	}
	clientID := make([]byte, 8)
	corerand.SecureRand.Read(clientID)

	cfg := &OAuthClientConfig{
		ClientID:        hex.EncodeToString(clientID),
		Name:            opts.Name,
		ApplicationType: opts.ApplicationType,
		RedirectURIs:    []string{opts.RedirectURI},
	}

	if opts.PostLogoutRedirectURI != "" {
		cfg.PostLogoutRedirectURIs = []string{opts.PostLogoutRedirectURI}
	}

	return cfg, nil
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
		Data: &CSRFKeyMaterials{Set: wrapInSet(secrets.GenerateOctetKeyForSig(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  WebhookKeyMaterialsKey,
		Data: &WebhookKeyMaterials{Set: wrapInSet(secrets.GenerateOctetKeyForSig(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  AdminAPIAuthKeyKey,
		Data: &AdminAPIAuthKey{Set: wrapInSet(secrets.GenerateRSAKey(createdAt, rng))},
	})

	items = append(items, SecretItem{
		Key:  ImagesKeyMaterialsKey,
		Data: &ImagesKeyMaterials{Set: wrapInSet(secrets.GenerateOctetKeyForSig(createdAt, rng))},
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
