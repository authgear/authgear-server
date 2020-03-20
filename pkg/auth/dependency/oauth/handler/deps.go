package handler

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideAuthorizationHandler(
	ctx context.Context,
	cfg *config.TenantConfiguration,
	lf logging.Factory,
	as oauth.AuthorizationStore,
	cs oauth.CodeGrantStore,
	authze oauth.AuthorizeEndpointProvider,
	authne oauth.AuthenticateEndpointProvider,
	vs ScopesValidator,
	cg TokenGenerator,
	tp time.Provider,
) *AuthorizationHandler {
	return &AuthorizationHandler{
		Context: ctx,
		AppID:   cfg.AppID,
		Clients: cfg.AppConfig.Clients,
		Logger:  lf.NewLogger("oauth-authz"),

		Authorizations:       as,
		CodeGrants:           cs,
		AuthorizeEndpoint:    authze,
		AuthenticateEndpoint: authne,
		ValidateScopes:       vs,
		CodeGenerator:        cg,
		Time:                 tp,
	}
}

func ProvideTokenHandler(
	ctx context.Context,
	cfg *config.TenantConfiguration,
	lf logging.Factory,
	as oauth.AuthorizationStore,
	cs oauth.CodeGrantStore,
	os oauth.OfflineGrantStore,
	ags oauth.AccessGrantStore,
	sp session.Provider,
	ti IDTokenIssuer,
	cg TokenGenerator,
	tp time.Provider,
) *TokenHandler {
	return &TokenHandler{
		Context: ctx,
		Clients: cfg.AppConfig.Clients,
		Logger:  lf.NewLogger("oauth-token"),

		Authorizations: as,
		CodeGrants:     cs,
		OfflineGrants:  os,
		AccessGrants:   ags,
		Sessions:       sp,
		IDTokenIssuer:  ti,
		GenerateToken:  cg,
		Time:           tp,
	}
}

var DependencySet = wire.NewSet(
	ProvideAuthorizationHandler,
	ProvideTokenHandler,
	wire.Value(TokenGenerator(GenerateToken)),
)
