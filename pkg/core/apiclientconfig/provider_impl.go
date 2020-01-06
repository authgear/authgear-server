package apiclientconfig

import (
	"crypto/subtle"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type providerImpl struct {
	authContext  coreAuth.ContextGetter
	tenantConfig config.TenantConfiguration
}

func NewProvider(authContext coreAuth.ContextGetter, tenantConfig config.TenantConfiguration) Provider {
	return &providerImpl{
		authContext:  authContext,
		tenantConfig: tenantConfig,
	}
}

func (p *providerImpl) Get() (string, *config.APIClientConfiguration, bool) {
	accessKey := p.authContext.AccessKey()
	if accessKey.ClientID == "" {
		return "", nil, false
	}
	clientID := accessKey.ClientID
	c, ok := model.GetClientConfig(p.tenantConfig.AppConfig.Clients, clientID)
	return clientID, c, ok
}

func (p *providerImpl) GetAccessKeyByAPIKey(apiKey string) model.AccessKey {
	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(p.tenantConfig.AppConfig.MasterKey)) == 1 {
		return model.AccessKey{Type: model.MasterAccessKeyType}
	}

	for _, clientConfig := range p.tenantConfig.AppConfig.Clients {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(clientConfig.APIKey)) == 1 {
			return model.AccessKey{Type: model.APIAccessKeyType, ClientID: clientConfig.ID}
		}
	}

	return model.AccessKey{Type: model.NoAccessKeyType}
}

func (p *providerImpl) GetAccessKeyByClientID(clientID string) model.AccessKey {
	return model.AccessKey{Type: model.APIAccessKeyType, ClientID: clientID}
}

var (
	_ Provider = &providerImpl{}
)
