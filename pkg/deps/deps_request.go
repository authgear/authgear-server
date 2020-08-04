package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/forgotpassword"
	identityanonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	identityservice "github.com/authgear/authgear-server/pkg/auth/dependency/identity/service"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/handler"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/auth/dependency/oidc/handler"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/dependency/verification"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	handlerinternal "github.com/authgear/authgear-server/pkg/auth/handler/internalserver"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	viewmodelswebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/middlewares"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var requestDeps = wire.NewSet(
	commonDeps,

	webapp.DependencySet,
	wire.Bind(new(oauthhandler.WebAppAuthenticateURLProvider), new(*webapp.AuthenticateURLProvider)),
	wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
	wire.Bind(new(sso.RedirectURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(forgotpassword.URLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(verification.WebAppURLProvider), new(*webapp.URLProvider)),

	oauthhandler.DependencySet,
	oidchandler.DependencySet,

	middlewares.DependencySet,

	handlerinternal.DependencySet,
	wire.Bind(new(handlerinternal.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
	wire.Bind(new(handlerinternal.VerificationService), new(*verification.Service)),

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	ProvideOAuthMetadataProviders,

	viewmodelswebapp.DependencySet,

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.PasswordPolicy), new(*password.Checker)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*auth.SessionManager)),
	wire.Bind(new(handlerwebapp.WebAppService), new(*webapp.Service)),
)
