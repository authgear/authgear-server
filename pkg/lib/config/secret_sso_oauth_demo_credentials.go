package config

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

var _ = SecretConfigSchema.Add("SSOOAuthDemoCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": {
			"type": "array",
			"items": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"provider_config": {
						"type": "object",
						"additionalProperties": true
					},
					"client_secret": {
						"type": "string"
					}
				},
				"required": ["provider_config", "client_secret"]
			}
		}
	},
	"required": ["items"]
}
`)

type SSOOAuthDemoCredentials struct {
	Items []SSOOAuthDemoCredentialsItems `json:"items,omitempty"`
}

type SSOOAuthDemoCredentialsItems struct {
	ProviderConfig oauthrelyingparty.ProviderConfig `json:"provider_config,omitempty"`
	ClientSecret   string                           `json:"client_secret,omitempty"`
}

func (c *SSOOAuthDemoCredentials) SensitiveStrings() []string {
	var secrets []string
	for _, item := range c.Items {
		if item.ClientSecret != "" {
			secrets = append(secrets, item.ClientSecret)
		}
	}
	return secrets
}

func (c *SSOOAuthDemoCredentials) LookupByProviderType(providerType string) (*SSOOAuthDemoCredentialsItems, bool) {
	for _, item := range c.Items {
		if item.ProviderConfig.Type() == providerType {
			return &item, true
		}
	}
	return nil, false
}
