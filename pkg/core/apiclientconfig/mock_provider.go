package apiclientconfig

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type MockProvider struct {
	ClientID        string
	APIClientConfig *config.APIClientConfiguration
}

func NewMockProvider(clientID string) *MockProvider {
	return &MockProvider{
		ClientID: clientID,
		APIClientConfig: &config.APIClientConfiguration{
			ClientName:           clientID,
			ClientID:             clientID,
			SessionTransport:     config.SessionTransportTypeHeader,
			AccessTokenLifetime:  1800,
			RefreshTokenLifetime: 86400,
			SameSite:             config.SessionCookieSameSiteLax,
		},
	}
}

func (p *MockProvider) Get() (string, *config.APIClientConfiguration, bool) {
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
