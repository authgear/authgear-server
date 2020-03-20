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

func ProvideAuthenticateProvider(
	validateProvider ValidateProvider,
	renderProvider RenderProvider,
	authnProvider AuthnProvider,
) AuthenticateProvider {
	return &AuthenticateProviderImpl{
		ValidateProvider: validateProvider,
		RenderProvider:   renderProvider,
		AuthnProvider:    authnProvider,
	}
}

var DependencySet = wire.NewSet(
	ProvideValidateProvider,
	ProvideAuthenticateProvider,
)

func ProvideCSPMiddleware(tConfig *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &CSPMiddleware{Clients: tConfig.AppConfig.Clients}
	return m.Handle
}
