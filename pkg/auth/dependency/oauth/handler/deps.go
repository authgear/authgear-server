package handler

import (
	"context"
	"net/http"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
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
	authze AuthorizeURLProvider,
	authne AuthenticateURLProvider,
	vs ScopesValidator,
	cg TokenGenerator,
	tp time.Provider,
) *AuthorizationHandler {
	return &AuthorizationHandler{
		Context: ctx,
		AppID:   cfg.AppID,
		Clients: cfg.AppConfig.Clients,
		Logger:  lf.NewLogger("oauth-authz"),

		Authorizations:  as,
		CodeGrants:      cs,
		AuthorizeURL:    authze,
		AuthenticateURL: authne,
		ValidateScopes:  vs,
		CodeGenerator:   cg,
		Time:            tp,
	}
}

func ProvideTokenHandler(
	r *http.Request,
	cfg *config.TenantConfiguration,
	lf logging.Factory,
	as oauth.AuthorizationStore,
	cs oauth.CodeGrantStore,
	os oauth.OfflineGrantStore,
	ags oauth.AccessGrantStore,
	aep auth.AccessEventProvider,
	sp session.Provider,
	aif AnonymousInteractionFlow,
	ti IDTokenIssuer,
	cg TokenGenerator,
	tp time.Provider,
) *TokenHandler {
	return &TokenHandler{
		Request: r,
		AppID:   cfg.AppID,
		Clients: cfg.AppConfig.Clients,
		Logger:  lf.NewLogger("oauth-token"),

		Authorizations: as,
		CodeGrants:     cs,
		OfflineGrants:  os,
		AccessGrants:   ags,
		AccessEvents:   aep,
		Sessions:       sp,
		Anonymous:      aif,
		IDTokenIssuer:  ti,
		GenerateToken:  cg,
		Time:           tp,
	}
}

var DependencySet = wire.NewSet(
	ProvideAuthorizationHandler,
	ProvideTokenHandler,
	wire.Struct(new(RevokeHandler), "*"),
	wire.Value(TokenGenerator(oauth.GenerateToken)),
	wire.Bind(new(authn.TokenIssuer), new(*TokenHandler)),
	wire.Bind(new(interactionflows.AuthAPITokenIssuer), new(*TokenHandler)),
	wire.Struct(new(URLProvider), "*"),
	wire.Bind(new(AuthorizeURLProvider), new(*URLProvider)),
)
