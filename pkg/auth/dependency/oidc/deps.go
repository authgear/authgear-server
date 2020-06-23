package oidc

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideIDTokenIssuer(
	cfg *config.TenantConfiguration,
	ep EndpointsProvider,
	u UserProvider,
	t clock.Clock,
) *IDTokenIssuer {
	return &IDTokenIssuer{
		OIDCConfig: *cfg.AppConfig.OIDC,
		Endpoints:  ep,
		Users:      u,
		Clock:      t,
	}
}

var DependencySet = wire.NewSet(
	wire.Value(handler.ScopesValidator(ValidateScopes)),
	wire.Struct(new(MetadataProvider), "*"),
	ProvideIDTokenIssuer,
	wire.Bind(new(handler.IDTokenIssuer), new(*IDTokenIssuer)),
)
