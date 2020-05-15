package oidc

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideIDTokenIssuer(
	cfg *config.TenantConfiguration,
	up urlprefix.Provider,
	as authinfo.Store,
	ups userprofile.Store,
	t time.Provider,
) *IDTokenIssuer {
	return &IDTokenIssuer{
		OIDCConfig:       *cfg.AppConfig.OIDC,
		URLPrefix:        up,
		AuthInfoStore:    as,
		UserProfileStore: ups,
		Time:             t,
	}
}

var DependencySet = wire.NewSet(
	wire.Value(handler.ScopesValidator(ValidateScopes)),
	wire.Struct(new(MetadataProvider), "*"),
	ProvideIDTokenIssuer,
	wire.Bind(new(handler.IDTokenIssuer), new(*IDTokenIssuer)),
)
