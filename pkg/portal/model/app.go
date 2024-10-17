package model

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type AppListItem struct {
	AppID        string `json:"appID,omitempty"`
	PublicOrigin string `json:"publicOrigin,omitempty"`
}

type App struct {
	ID      string
	Context *config.AppContext
}

type AppResource struct {
	DescriptedPath resource.DescriptedPath
	Context        *config.AppContext
}

type WebhookSecret struct {
	Secret *string `json:"secret,omitempty"`
}

type OAuthSSOProviderClientSecret struct {
	Alias        string  `json:"alias,omitempty"`
	ClientSecret *string `json:"clientSecret,omitempty"`
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

type BotProtectionProviderSecret struct {
	Type      config.BotProtectionProviderType `json:"type,omitempty"`
	SecretKey *string                          `json:"secretKey,omitempty"`
}

type SAMLIdpSigningCertificate struct {
	KeyID                  string `json:"keyID,omitempty"`
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	CertificatePEM         string `json:"certificatePEM,omitempty"`
}

type SAMLIdpSigningSecrets struct {
	Certificates []SAMLIdpSigningCertificate `json:"certificates,omitempty"`
}

type SAMLSpSigningCertificate struct {
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	CertificatePEM         string `json:"certificatePEM,omitempty"`
}

type SAMLSpSigningSecrets struct {
	ClientID     string                     `json:"clientID,omitempty"`
	Certificates []SAMLSpSigningCertificate `json:"certificates,omitempty"`
}

type SecretConfig struct {
	OAuthSSOProviderClientSecrets []OAuthSSOProviderClientSecret `json:"oauthSSOProviderClientSecrets,omitempty"`
	WebhookSecret                 *WebhookSecret                 `json:"webhookSecret,omitempty"`
	AdminAPISecrets               []AdminAPISecret               `json:"adminAPISecrets,omitempty"`
	SMTPSecret                    *SMTPSecret                    `json:"smtpSecret,omitempty"`
	OAuthClientSecrets            []OAuthClientSecret            `json:"oauthClientSecrets,omitempty"`
	BotProtectionProviderSecret   *BotProtectionProviderSecret   `json:"botProtectionProviderSecret,omitempty"`
	SAMLIdpSigningSecrets         *SAMLIdpSigningSecrets         `json:"samlIdpSigningSecrets,omitempty"`
	SAMLSpSigningSecrets          []SAMLSpSigningSecrets         `json:"samlSpSigningSecrets,omitempty"`
}

//nolint:gocognit
func NewSecretConfig(secretConfig *config.SecretConfig, unmaskedSecrets []config.SecretKey, now time.Time) (*SecretConfig, error) {
	out := &SecretConfig{}
	var unmaskedSecretsSet map[config.SecretKey]interface{} = map[config.SecretKey]interface{}{}
	for _, s := range unmaskedSecrets {
		unmaskedSecretsSet[s] = s
	}

	if oauthSSOProviderCredentials, ok := secretConfig.LookupData(config.OAuthSSOProviderCredentialsKey).(*config.OAuthSSOProviderCredentials); ok {
		for _, item := range oauthSSOProviderCredentials.Items {
			var clientSecret *string = nil
			if _, exist := unmaskedSecretsSet[config.OAuthSSOProviderCredentialsKey]; exist {
				s := item.ClientSecret
				clientSecret = &s
			}
			out.OAuthSSOProviderClientSecrets = append(out.OAuthSSOProviderClientSecrets, OAuthSSOProviderClientSecret{
				Alias:        item.Alias,
				ClientSecret: clientSecret,
			})
		}
	}

	if webhook, ok := secretConfig.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials); ok {
		if webhook.Set.Len() == 1 {
			if jwkKey, ok := webhook.Set.Key(0); ok {
				if sKey, ok := jwkKey.(jwk.SymmetricKey); ok {
					var secret *string
					if _, exist := unmaskedSecretsSet[config.WebhookKeyMaterialsKey]; exist {
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
			if jwkKey, ok := adminAPI.Set.Key(i); ok {
				var createdAt *time.Time
				if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
					if fCreatedAt, ok := anyCreatedAt.(float64); ok {
						t := time.Unix(int64(fCreatedAt), 0).UTC()
						createdAt = &t
					}
				}
				set := jwk.NewSet()
				_ = set.AddKey(jwkKey)
				publicKeyPEMBytes, err := jwkutil.PublicPEM(set)
				if err != nil {
					return nil, err
				}

				var privateKeyPEM *string
				if _, exist := unmaskedSecretsSet[config.AdminAPIAuthKeyKey]; exist {
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
		if _, exist := unmaskedSecretsSet[config.SMTPServerCredentialsKey]; exist {
			smtpSecret.Password = &smtp.Password
		}
		out.SMTPSecret = smtpSecret
	}

	if oauthClientSecrets, ok := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials); ok {
		for _, item := range oauthClientSecrets.Items {

			keys := []OAuthClientSecretKey{}

			for i := 0; i < item.Set.Len(); i++ {
				if jwkKey, ok := item.Set.Key(i); ok {
					newlyCreated := false
					var createdAt *time.Time
					if anyCreatedAt, ok := jwkKey.Get("created_at"); ok {
						if fCreatedAt, ok := anyCreatedAt.(float64); ok {
							t := time.Unix(int64(fCreatedAt), 0).UTC()
							createdAt = &t
							elapsed := now.Sub(*createdAt)
							newlyCreated = elapsed < 5*time.Minute
						}
					}
					var bytes []byte
					err := jwkKey.Raw(&bytes)
					if err != nil {
						return nil, err
					}

					clientSecret := ""
					_, unmask := unmaskedSecretsSet[config.OAuthClientCredentialsKey]
					if unmask || newlyCreated {
						clientSecret = string(bytes)
					}

					keys = append(keys, OAuthClientSecretKey{
						KeyID:     jwkKey.KeyID(),
						Key:       clientSecret,
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

	if botProtectionProviderSecrets, ok := secretConfig.LookupData(config.BotProtectionProviderCredentialsKey).(*config.BotProtectionProviderCredentials); ok {
		bpSecret := &BotProtectionProviderSecret{
			Type: botProtectionProviderSecrets.Type,
		}

		if _, exist := unmaskedSecretsSet[config.BotProtectionProviderCredentialsKey]; exist {
			bpSecret.SecretKey = &botProtectionProviderSecrets.SecretKey
		}

		out.BotProtectionProviderSecret = bpSecret
	}

	if samlIdpSigningSecrets, ok := secretConfig.LookupData(config.SAMLIdpSigningMaterialsKey).(*config.SAMLIdpSigningMaterials); ok {
		out.SAMLIdpSigningSecrets = toPortalSAMLIdpSigningSecrets(samlIdpSigningSecrets)
	}

	if samlSpSigningSecrets, ok := secretConfig.LookupData(config.SAMLSpSigningMaterialsKey).(*config.SAMLSpSigningMaterials); ok {
		out.SAMLSpSigningSecrets = toPortalSAMLSpSigningSecrets(samlSpSigningSecrets)
	}

	return out, nil
}

func toPortalSAMLIdpSigningSecrets(cfg *config.SAMLIdpSigningMaterials) *SAMLIdpSigningSecrets {
	result := &SAMLIdpSigningSecrets{
		Certificates: []SAMLIdpSigningCertificate{},
	}

	for _, cfgCert := range cfg.Certificates {
		cert := SAMLIdpSigningCertificate{
			KeyID:                  cfgCert.Key.KeyID(),
			CertificateFingerprint: cfgCert.Certificate.Fingerprint(),
			CertificatePEM:         string(cfgCert.Certificate.Pem),
		}
		result.Certificates = append(result.Certificates, cert)
	}

	return result
}

func toPortalSAMLSpSigningSecrets(cfg *config.SAMLSpSigningMaterials) []SAMLSpSigningSecrets {
	result := []SAMLSpSigningSecrets{}

	for _, cfgItem := range *cfg {
		item := SAMLSpSigningSecrets{
			ClientID:     cfgItem.ServiceProviderID,
			Certificates: []SAMLSpSigningCertificate{},
		}

		for _, cfgCert := range cfgItem.Certificates {
			item.Certificates = append(item.Certificates, SAMLSpSigningCertificate{
				CertificateFingerprint: cfgCert.Fingerprint(),
				CertificatePEM:         string(cfgCert.Pem),
			})
		}

		result = append(result, item)
	}

	return result
}
