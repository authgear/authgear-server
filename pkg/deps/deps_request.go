package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/forgotpassword"
	identityanonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	identityprovider "github.com/authgear/authgear-server/pkg/auth/dependency/identity/provider"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/handler"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/auth/dependency/oidc/handler"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	handlerinternal "github.com/authgear/authgear-server/pkg/auth/handler/internalserver"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/middlewares"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var requestDeps = wire.NewSet(
	wire.NewSet(
		commonDeps,

		wire.NewSet(
			sso.DependencySet,
			wire.Bind(new(webapp.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
			wire.Bind(new(webapp.SSOStateCodec), new(*sso.StateCodec)),
		),

		wire.NewSet(
			webapp.DependencySet,
			wire.Bind(new(oauthhandler.WebAppURLProvider), new(*webapp.URLProvider)),
			wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
			wire.Bind(new(sso.RedirectURLProvider), new(*webapp.URLProvider)),
			wire.Bind(new(forgotpassword.URLProvider), new(*webapp.URLProvider)),
		),

		oauthhandler.DependencySet,
		oidchandler.DependencySet,
	),

	middlewares.DependencySet,

	handlerinternal.DependencySet,
	wire.Bind(new(handlerinternal.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	ProvideOAuthMetadataProviders,

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.LoginOAuthService), new(*webapp.OAuthService)),
	// wire.Bind(new(handlerwebapp.SignupProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.PromoteProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.SSOProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.EnterLoginIDProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.EnterPasswordProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.CreatePasswordProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.OOBOTPProvider), new(*webapp.AuthenticateProviderImpl)),
	// wire.Bind(new(handlerwebapp.ForgotPasswordProvider), new(*webapp.ForgotPasswordProvider)),
	// wire.Bind(new(handlerwebapp.ForgotPasswordSuccessProvider), new(*webapp.ForgotPasswordProvider)),
	// wire.Bind(new(handlerwebapp.ResetPasswordProvider), new(*webapp.ForgotPasswordProvider)),
	// wire.Bind(new(handlerwebapp.ResetPasswordSuccessProvider), new(*webapp.ForgotPasswordProvider)),
	// wire.Bind(new(handlerwebapp.SettingsIdentityProvider), new(*webapp.AuthenticateProviderImpl)),

	wire.Bind(new(handlerwebapp.IdentityProvider), new(*identityprovider.Provider)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*auth.SessionManager)),
)
