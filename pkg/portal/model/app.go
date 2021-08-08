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

type SMTPSecret struct {
	Host     string  `json:"host,omitempty"`
	Port     int     `json:"port,omitempty"`
	Username string  `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

type SecretConfig struct {
	OAuthClientSecrets []OAuthClientSecret `json:"oauthClientSecrets,omitempty"`
	WebhookSecret      *WebhookSecret      `json:"webhookSecret,omitempty"`
	AdminAPISecrets    []AdminAPISecret    `json:"adminAPISecrets,omitempty"`
	SMTPSecret         *SMTPSecret         `json:"smtpSecret,omitempty"`
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

	if smtp, ok := secretConfig.LookupData(config.SMTPServerCredentialsKey).(*config.SMTPServerCredentials); ok {
		smtpSecret := &SMTPSecret{
			Host:     smtp.Host,
			Port:     smtp.Port,
			Username: smtp.Username,
		}
		if unmasked {
			smtpSecret.Password = &smtp.Password
		}
		out.SMTPSecret = smtpSecret
	}

	return out, nil
}

func (c *SecretConfig) updateOAuthClientCredentials() (item *config.SecretItem, err error) {
	if len(c.OAuthClientSecrets) <= 0 {
		return
	}

	// The strategy is simply use incoming one.
	var oauthItems []config.OAuthClientCredentialsItem
	for _, secret := range c.OAuthClientSecrets {
		oauthItems = append(oauthItems, config.OAuthClientCredentialsItem{
			Alias:        secret.Alias,
			ClientSecret: secret.ClientSecret,
		})
	}

	var data []byte
	data, err = json.Marshal(&config.OAuthClientCredentials{
		Items: oauthItems,
	})
	if err != nil {
		return
	}

	item = &config.SecretItem{
		Key:     config.OAuthClientCredentialsKey,
		RawData: json.RawMessage(data),
	}
	return
}

func (c *SecretConfig) updateSMTP(currentConfig *config.SecretConfig) (item *config.SecretItem, err error) {
	if c.SMTPSecret == nil {
		return
	}

	_, existingItem, ok := currentConfig.Lookup(config.SMTPServerCredentialsKey)

	if c.SMTPSecret.Password == nil {
		// No change
		if ok {
			item = existingItem
		}
	} else {
		var data []byte
		data, err = json.Marshal(&config.SMTPServerCredentials{
			Host:     c.SMTPSecret.Host,
			Port:     c.SMTPSecret.Port,
			Username: c.SMTPSecret.Username,
			Password: *c.SMTPSecret.Password,
		})
		if err != nil {
			return
		}

		item = &config.SecretItem{
			Key:     config.SMTPServerCredentialsKey,
			RawData: json.RawMessage(data),
		}
	}

	return
}

func (c *SecretConfig) ToYAMLForUpdate(currentConfig *config.SecretConfig) ([]byte, error) {
	var items []config.SecretItem

	oauthItem, err := c.updateOAuthClientCredentials()
	if err != nil {
		return nil, err
	}
	if oauthItem != nil {
		items = append(items, *oauthItem)
	}

	smtpItem, err := c.updateSMTP(currentConfig)
	if err != nil {
		return nil, err
	}
	if smtpItem != nil {
		items = append(items, *smtpItem)
	}

	newConfig := &config.SecretConfig{
		Secrets: items,
	}

	jsonBytes, err := json.Marshal(newConfig)
	if err != nil {
		return nil, err
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return nil, err
	}

	return yamlBytes, nil
}
