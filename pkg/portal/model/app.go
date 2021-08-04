package model

import (
	"encoding/json"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
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

type AdminAPISecret struct {
	KeyID         string     `json:"keyID,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	PublicKeyPEM  string     `json:"publicKeyPEM,omitempty"`
	PrivateKeyPEM *string    `json:"privateKeyPEM,omitempty"`
}

type SecretConfig struct {
	OAuthClientSecrets []OAuthClientSecret `json:"oauthClientSecrets,omitempty"`
	WebhookSecret      *WebhookSecret      `json:"webhookSecret,omitempty"`
	AdminAPISecrets    []AdminAPISecret    `json:"adminAPISecrets,omitempty"`
}

func NewSecretConfig(secretConfig *config.SecretConfig, unmasked bool) (*SecretConfig, error) {
	out := &SecretConfig{}

	if oauthClientCredentials, ok := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials); ok {
		for _, item := range oauthClientCredentials.Items {
			out.OAuthClientSecrets = append(out.OAuthClientSecrets, OAuthClientSecret{
				Alias:        item.Alias,
				ClientSecret: item.ClientSecret,
			})
		}
	}

	if webhook, ok := secretConfig.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials); ok {
		if webhook.Set.Len() == 1 {
			if jwkKey, ok := webhook.Set.Get(0); ok {
				if sKey, ok := jwkKey.(jwk.SymmetricKey); ok {
					var secret *string
					if unmasked {
						octets := sKey.Octets()
						octetsStr := string(octets)
						secret = &octetsStr
					}
					out.WebhookSecret = &WebhookSecret{
						Secret: secret,
					}
				}
			}
		}
	}

	if adminAPI, ok := secretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey); ok {
		for i := 0; i < adminAPI.Set.Len(); i++ {
			if jwkKey, ok := adminAPI.Set.Get(i); ok {
				var createdAt *time.Time
				if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
					if fCreatedAt, ok := anyCreatedAt.(float64); ok {
						t := time.Unix(int64(fCreatedAt), 0).UTC()
						createdAt = &t
					}
				}
				set := jwk.NewSet()
				_ = set.Add(jwkKey)
				publicKeyPEMBytes, err := jwkutil.PublicPEM(set)
				if err != nil {
					return nil, err
				}

				var privateKeyPEM *string
				if unmasked {
					privateKeyPEMBytes, err := jwkutil.PrivatePublicPEM(set)
					if err != nil {
						return nil, err
					}
					privateKeyPEMStr := string(privateKeyPEMBytes)
					privateKeyPEM = &privateKeyPEMStr
				}

				out.AdminAPISecrets = append(out.AdminAPISecrets, AdminAPISecret{
					KeyID:         jwkKey.KeyID(),
					CreatedAt:     createdAt,
					PublicKeyPEM:  string(publicKeyPEMBytes),
					PrivateKeyPEM: privateKeyPEM,
				})
			}
		}
	}

	return out, nil
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
