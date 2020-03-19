package handler

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

// TODO(oauth): implement stores
func ProvideAuthorizationHandler(
	ctx context.Context,
	cfg *config.TenantConfiguration,
	//as oauth.AuthorizationStore,
	//cs oauth.CodeGrantStore,
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

		//Authorizations:       as,
		//CodeGrants:           cs,
		AuthorizeEndpoint:    authze,
		AuthenticateEndpoint: authne,
		ValidateScopes:       vs,
		CodeGenerator:        cg,
		Time:                 tp,
	}
}

var DependencySet = wire.NewSet(
	ProvideAuthorizationHandler,
	wire.Value(TokenGenerator(GenerateToken)),
)
