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

func (p *providerImpl) Get() (*config.APIClientConfiguration, bool) {
	accessKey := p.authContext.AccessKey()
	if accessKey.ClientID == "" {
		return nil, false
	}
	return model.GetClientConfig(p.tenantConfig.UserConfig.Clients, accessKey.ClientID)
}

func (p *providerImpl) AccessKey(apiKey string) model.AccessKey {
	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(p.tenantConfig.UserConfig.MasterKey)) == 1 {
		return model.AccessKey{Type: model.MasterAccessKeyType}
	}

	for id, clientConfig := range p.tenantConfig.UserConfig.Clients {
		if clientConfig.Disabled {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(clientConfig.APIKey)) == 1 {
			return model.AccessKey{Type: model.APIAccessKeyType, ClientID: id}
		}
	}

	return model.AccessKey{Type: model.NoAccessKeyType}
}

var (
	_ Provider = &providerImpl{}
)
