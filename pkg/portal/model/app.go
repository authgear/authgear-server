package model

import (
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

type StructuredSecretConfig struct {
	OAuthClientSecrets []OAuthClientSecret `json:"oauthClientSecrets,omitempty"`
}

type OAuthClientSecret struct {
	Alias        string `json:"alias,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

func NewStructuredSecretConfig(secretConfig *config.SecretConfig) *StructuredSecretConfig {
	out := &StructuredSecretConfig{}

	oauthClientCredentials, _ := secretConfig.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials)

	for _, item := range oauthClientCredentials.Items {
		out.OAuthClientSecrets = append(out.OAuthClientSecrets, OAuthClientSecret{
			Alias:        item.Alias,
			ClientSecret: item.ClientSecret,
		})
	}

	return out
}
