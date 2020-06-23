package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	identityanonymous "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	oidchandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	handleroauth "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	handlersession "github.com/skygeario/skygear-server/pkg/auth/handler/session"
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var requestDeps = wire.NewSet(
	wire.NewSet(
		commonDeps,

		wire.NewSet(
			webapp.DependencySet,
			wire.Bind(new(oauthhandler.WebAppURLProvider), new(*webapp.URLProvider)),
			wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
		),

		oauthhandler.DependencySet,
		oidchandler.DependencySet,
	),

	middlewares.DependencySet,

	handlersession.DependencySet,
	wire.Bind(new(handlersession.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	ProvideOAuthMetadataProviders,
)
