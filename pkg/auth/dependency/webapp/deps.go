package webapp

import (
	"github.com/google/wire"

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
) AuthenticateProvider {
	return &AuthenticateProviderImpl{
		ValidateProvider: validateProvider,
		RenderProvider:   renderProvider,
	}
}

var DependencySet = wire.NewSet(
	ProvideValidateProvider,
	ProvideAuthenticateProvider,
)
