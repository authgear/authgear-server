package webapp

import (
	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideValidateProvider(tConfig *config.TenantConfiguration) ValidateProvider {
	return &ValidateProviderImpl{
		Validator:         validator,
		AuthConfiguration: tConfig.AppConfig.Auth,
	}
}

var DependencySet = wire.NewSet(
	ProvideValidateProvider,
	wire.Struct(new(StateStoreImpl), "*"),
	wire.Bind(new(StateStore), new(*StateStoreImpl)),
)

func ProvideCSPMiddleware(tConfig *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &CSPMiddleware{Clients: tConfig.AppConfig.Clients}
	return m.Handle
}

func ProvideStateMiddleware(stateStore StateStore) mux.MiddlewareFunc {
	m := &StateMiddleware{StateStore: stateStore}
	return m.Handle
}
