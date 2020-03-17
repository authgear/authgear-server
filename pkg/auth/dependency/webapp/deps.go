package webapp

import (
	"github.com/google/wire"
)

func ProvideValidateProvider() ValidateProvider {
	return &ValidateProviderImpl{
		Validator: validator,
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
