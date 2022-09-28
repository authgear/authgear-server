package model

import (
	"time"

	"github.com/lestrrat-go/jwx/jwk"

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

type OAuthSSOProviderClientSecret struct {
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

type OAuthClientSecretKey struct {
	KeyID     string     `json:"keyID,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Key       string     `json:"key,omitempty"`
}

type OAuthClientSecret struct {
	ClientID string                 `json:"clientID,omitempty"`
	Keys     []OAuthClientSecretKey `json:"keys,omitempty"`
}

type SecretConfig struct {
	OAuthSSOProviderClientSecrets []OAuthSSOProviderClientSecret `json:"oauthSSOProviderClientSecrets,omitempty"`
	WebhookSecret                 *WebhookSecret                 `json:"webhookSecret,omitempty"`
	AdminAPISecrets               []AdminAPISecret               `json:"adminAPISecrets,omitempty"`
	SMTPSecret                    *SMTPSecret                    `json:"smtpSecret,omitempty"`
	OAuthClientSecrets            []OAuthClientSecret            `json:"oauthClientSecrets,omitempty"`
}

func NewSecretConfig(secretConfig *config.SecretConfig, unmasked bool) (*SecretConfig, error) {
	out := &SecretConfig{}

	if oauthSSOProviderCredentials, ok := secretConfig.LookupData(config.OAuthSSOProviderCredentialsKey).(*config.OAuthSSOProviderCredentials); ok {
		for _, item := range oauthSSOProviderCredentials.Items {
			out.OAuthSSOProviderClientSecrets = append(out.OAuthSSOProviderClientSecrets, OAuthSSOProviderClientSecret{
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

	if oauthClientSecrets, ok := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials); ok {
		for _, item := range oauthClientSecrets.Items {

			keys := []OAuthClientSecretKey{}

			for i := 0; i < item.Set.Len(); i++ {
				if jwkKey, ok := item.Set.Get(i); ok {
					var createdAt *time.Time
					if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
						if fCreatedAt, ok := anyCreatedAt.(float64); ok {
							t := time.Unix(int64(fCreatedAt), 0).UTC()
							createdAt = &t
						}
					}
					var bytes []byte
					err := jwkKey.Raw(&bytes)
					if err != nil {
						return nil, err
					}

					keys = append(keys, OAuthClientSecretKey{
						KeyID:     jwkKey.KeyID(),
						Key:       string(bytes),
						CreatedAt: createdAt,
					})
				}
			}

			out.OAuthClientSecrets = append(out.OAuthClientSecrets, OAuthClientSecret{
				ClientID: item.ClientID,
				Keys:     keys,
			})
		}
	}

	return out, nil
}
