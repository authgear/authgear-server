package model

import (
	"encoding/json"

	"github.com/lestrrat-go/jwx/jwk"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
)

type App struct {
	ID      string
	Context *config.AppContext
}

type AppResource struct {
	DescriptedPath appresource.DescriptedPath
	Context        *config.AppContext
}

type WebhookSecret struct {
	Secret *string `json:"secret,omitempty"`
}

type OAuthClientSecret struct {
	Alias        string `json:"alias,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type SecretConfig struct {
	OAuthClientSecrets []OAuthClientSecret `json:"oauthClientSecrets,omitempty"`
	WebhookSecret      *WebhookSecret      `json:"webhookSecret,omitempty"`
}

func NewSecretConfig(secretConfig *config.SecretConfig) *SecretConfig {
	out := &SecretConfig{}

	oauthClientCredentials, ok := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials)
	if ok {
		for _, item := range oauthClientCredentials.Items {
			out.OAuthClientSecrets = append(out.OAuthClientSecrets, OAuthClientSecret{
				Alias:        item.Alias,
				ClientSecret: item.ClientSecret,
			})
		}
	}

	webhook, ok := secretConfig.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials)
	if ok {
		if webhook.Set.Len() == 1 {
			if jwkKey, ok := webhook.Set.Get(0); ok {
				if _, ok := jwkKey.(jwk.SymmetricKey); ok {
					// octets := sKey.Octets()
					out.WebhookSecret = &WebhookSecret{
						// FIXME(secret): expose unmasked secret.
					}
				}
			}
		}
	}

	return out
}

type secretItem struct {
	Key  string      `json:"key,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (c *SecretConfig) ToYAMLForUpdate() ([]byte, error) {
	var items []secretItem
	if c.OAuthClientSecrets != nil {
		var oauthItems []config.OAuthClientCredentialsItem
		for _, secret := range c.OAuthClientSecrets {
			oauthItems = append(oauthItems, config.OAuthClientCredentialsItem{
				Alias:        secret.Alias,
				ClientSecret: secret.ClientSecret,
			})
		}

		items = append(items, secretItem{
			Key: string(config.OAuthClientCredentialsKey),
			Data: &config.OAuthClientCredentials{
				Items: oauthItems,
			},
		})
	}

	jsonBytes, err := json.Marshal(map[string]interface{}{
		"secrets": items,
	})
	if err != nil {
		return nil, err
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return nil, err
	}

	return yamlBytes, nil
}
