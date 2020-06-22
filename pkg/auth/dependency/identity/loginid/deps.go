package loginid

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/db"
)

func ProvideTypeCheckerFactory(
	config *config.TenantConfiguration,
	reservedNameChecker *ReservedNameChecker,
) *TypeCheckerFactory {
	return &TypeCheckerFactory{
		Keys:                config.AppConfig.Identity.LoginID.Keys,
		Types:               config.AppConfig.Identity.LoginID.Types,
		ReservedNameChecker: reservedNameChecker,
	}
}

func ProvideChecker(
	config *config.TenantConfiguration,
	typeCheckerFactory *TypeCheckerFactory,
) *Checker {
	return &Checker{
		Keys:               config.AppConfig.Identity.LoginID.Keys,
		Types:              config.AppConfig.Identity.LoginID.Types,
		TypeCheckerFactory: typeCheckerFactory,
	}
}

func ProvideNormalizerFactory(
	config *config.TenantConfiguration,
) *NormalizerFactory {
	return &NormalizerFactory{
		Keys:  config.AppConfig.Identity.LoginID.Keys,
		Types: config.AppConfig.Identity.LoginID.Types,
	}
}

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t clock.Clock,
	c *config.TenantConfiguration,
	checker *Checker,
	normalizerFactory *NormalizerFactory,
) *Provider {
	config := *c.AppConfig.Identity.LoginID
	return &Provider{
		Store:             &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Config:            config,
		Checker:           checker,
		NormalizerFactory: normalizerFactory,
	}
}

var DependencySet = wire.NewSet(
	ProvideTypeCheckerFactory,
	ProvideChecker,
	ProvideNormalizerFactory,
	ProvideProvider,
)
