package loginid

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideLoginIDChecker(
	config *config.TenantConfiguration,
	reservedNameChecker *ReservedNameChecker,
) LoginIDChecker {
	return NewDefaultLoginIDChecker(
		config.AppConfig.Identity.LoginID.Keys,
		config.AppConfig.Identity.LoginID.Types,
		reservedNameChecker,
	)
}

func ProvideLoginIDNormalizerFactory(
	config *config.TenantConfiguration,
) LoginIDNormalizerFactory {
	return NewLoginIDNormalizerFactory(
		config.AppConfig.Identity.LoginID.Keys,
		config.AppConfig.Identity.LoginID.Types,
	)
}

var DependencySet = wire.NewSet(
	ProvideLoginIDChecker,
	ProvideLoginIDNormalizerFactory,
)
