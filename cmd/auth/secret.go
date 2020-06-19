package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
	"sigs.k8s.io/yaml"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

func GenerateKID() (kid string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return
	}

	kid = hex.EncodeToString(b)
	return
}

func GenerateSymetricKey() (data interface{}, err error) {
	key := make([]byte, 32)

	_, err = rand.Read(key)
	if err != nil {
		return
	}

	kid, err := GenerateKID()
	if err != nil {
		return
	}

	jwkKey, err := jwk.New(key)
	if err != nil {
		return
	}

	jwkKey.Set(jwk.KeyIDKey, kid)
	jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	jwkKey.Set(jwk.AlgorithmKey, "HS256")

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	data = keySet

	return
}

func GenerateRSAKeypair() (data interface{}, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	kid, err := GenerateKID()
	if err != nil {
		return
	}

	jwkKey, err := jwk.New(privateKey)
	if err != nil {
		return
	}
	jwkKey.Set(jwk.KeyIDKey, kid)
	jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	jwkKey.Set(jwk.AlgorithmKey, "RS256")

	keySet := jwk.Set{
		Keys: []jwk.Key{jwkKey},
	}
	data = keySet

	return
}

func PrintSecretYAML() {
	var items []config.SecretItem
	items = append(items, config.SecretItem{
		Key: config.DatabaseCredentialsKey,
		Data: config.DatabaseCredentials{
			DatabaseURL:    "postgres://postgres:@localhost:5432/postgres?sslmode=disable",
			DatabaseSchema: "public",
		},
	})
	items = append(items, config.SecretItem{
		Key: config.RedisCredentialsKey,
		Data: config.RedisCredentials{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	})
	items = append(items, config.SecretItem{
		Key: config.SMTPServerCredentialsKey,
		Data: config.SMTPServerCredentials{
			Host:     "localhost",
			Port:     587,
			Mode:     config.SMTPModeSSL,
			Username: "username",
			Password: "password",
		},
	})
	items = append(items, config.SecretItem{
		Key: config.TwilioCredentialsKey,
		Data: config.TwilioCredentials{
			AccountSID: "account_sid",
			AuthToken:  "auth_token",
		},
	})
	items = append(items, config.SecretItem{
		Key: config.NexmoCredentialsKey,
		Data: config.NexmoCredentials{
			APIKey:    "api_key",
			APISecret: "api_secret",
		},
	})

	jwt, _ := GenerateRSAKeypair()
	items = append(items, config.SecretItem{
		Key:  config.JWTKeyMaterialsKey,
		Data: jwt,
	})

	oidc, _ := GenerateRSAKeypair()
	items = append(items, config.SecretItem{
		Key:  config.OIDCKeyMaterialsKey,
		Data: oidc,
	})

	csrf, _ := GenerateSymetricKey()
	items = append(items, config.SecretItem{
		Key:  config.CSRFKeyMaterialsKey,
		Data: csrf,
	})

	webhook, _ := GenerateSymetricKey()
	items = append(items, config.SecretItem{
		Key:  config.WebhookKeyMaterialsKey,
		Data: webhook,
	})

	for idx, item := range items {
		rawMessage, _ := json.Marshal(item.Data)
		items[idx].RawData = json.RawMessage(rawMessage)
	}

	b, _ := yaml.Marshal(config.SecretConfig{
		Secrets: items,
	})

	fmt.Printf("%s", string(b))
}
