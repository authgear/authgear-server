package apiclientconfig

import (
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type providerImpl struct {
	authContext coreAuth.ContextGetter
	clients     map[string]config.APIClientConfiguration
}

func NewProvider(authContext coreAuth.ContextGetter, clients map[string]config.APIClientConfiguration) Provider {
	return &providerImpl{
		authContext: authContext,
		clients:     clients,
	}
}

func (p *providerImpl) Get() (*config.APIClientConfiguration, bool) {
	accessKey := p.authContext.AccessKey()
	if accessKey.ClientID == "" {
		return nil, false
	}
	return model.GetClientConfig(p.clients, accessKey.ClientID)
}

var (
	_ Provider = &providerImpl{}
)
