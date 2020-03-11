package apiclientconfig

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type MockProvider struct {
	ClientID        string
	APIClientConfig config.OAuthClientConfiguration
}

func NewMockProvider(clientID string) *MockProvider {
	return &MockProvider{
		ClientID: clientID,
		APIClientConfig: config.OAuthClientConfiguration{
			"client_name":            clientID,
			"client_id":              clientID,
			"access_token_lifetime":  1800.0,
			"refresh_token_lifetime": 86400.0,
		},
	}
}

func (p *MockProvider) Get() (string, config.OAuthClientConfiguration, bool) {
	if p.APIClientConfig == nil {
		return "", nil, false
	}
	return p.ClientID, p.APIClientConfig, true
}

func (p *MockProvider) GetAccessKeyByAPIKey(apiKey string) model.AccessKey {
	if p.APIClientConfig == nil {
		return model.AccessKey{Type: model.NoAccessKeyType}
	}
	return model.AccessKey{Type: model.APIAccessKeyType, ClientID: p.ClientID}
}

func (p *MockProvider) GetAccessKeyByClientID(clientID string) model.AccessKey {
	return model.AccessKey{Type: model.APIAccessKeyType, ClientID: clientID}
}

var (
	_ Provider = &MockProvider{}
)
