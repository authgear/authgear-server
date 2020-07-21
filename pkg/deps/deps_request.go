package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/forgotpassword"
	identityanonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	identityprovider "github.com/authgear/authgear-server/pkg/auth/dependency/identity/provider"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
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
	"github.com/authgear/authgear-server/pkg/middlewares"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var requestDeps = wire.NewSet(
	commonDeps,

	sso.DependencySet,
	wire.Bind(new(interactionflows.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),

	webapp.DependencySet,
	wire.Bind(new(webapp.ResponderStates), new(*interactionflows.StateService)),
	wire.Bind(new(webapp.URLProviderStates), new(*interactionflows.StateService)),
	wire.Bind(new(webapp.StateMiddlewareStates), new(*interactionflows.StateStoreRedis)),
	wire.Bind(new(oauthhandler.WebAppURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
	wire.Bind(new(sso.RedirectURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(forgotpassword.URLProvider), new(*webapp.URLProvider)),

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

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.StateService), new(*interactionflows.StateService)),
	wire.Bind(new(handlerwebapp.Responder), new(*webapp.Responder)),
	wire.Bind(new(handlerwebapp.EnterPasswordInteractions), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(handlerwebapp.ForgotPasswordInteractions), new(*forgotpassword.Provider)),
	wire.Bind(new(handlerwebapp.ResetPasswordInteractions), new(*forgotpassword.Provider)),
	wire.Bind(new(handlerwebapp.OOBOTPInteractions), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(handlerwebapp.SSOCallbackInteractions), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(handlerwebapp.SettingsIdentityInteractions), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(handlerwebapp.EnterLoginIDInteractions), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(handlerwebapp.PromoteInteractions), new(*interactionflows.WebAppFlow)),

	wire.Bind(new(handlerwebapp.IdentityProvider), new(*identityprovider.Provider)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*auth.SessionManager)),
)
