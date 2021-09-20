package auth

import (
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicEndMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newPanicLogMiddleware),
		p.Middleware(newPublicOriginMiddleware),
	)

	staticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		httproute.MiddlewareFunc(httputil.ETag),
	)

	oauthStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		httproute.MiddlewareFunc(httputil.ETag),
	)

	oauthAPIChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		p.Middleware(newCORSMiddleware),
		p.Middleware(newWebAppWeChatRedirectURIMiddleware),
	)

	apiChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		p.Middleware(newPanicAPIMiddleware),
		p.Middleware(newCORSMiddleware),
	)

	scopedChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		p.Middleware(newPanicWriteEmptyResponseMiddleware),
		p.Middleware(newCORSMiddleware),
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	webappChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		httproute.MiddlewareFunc(webapp.IntlMiddleware),
		p.Middleware(newPanicWebAppMiddleware),
		p.Middleware(newWebAppSessionMiddleware),
		p.Middleware(newWebAppUILocalesMiddleware),
		p.Middleware(newWebAppWeChatRedirectURIMiddleware),
		p.Middleware(newWebAppClientIDMiddleware),
	)
	webappSSOCallbackChain := httproute.Chain(
		webappChain,
	)
	webappWebsocketChain := httproute.Chain(
		webappChain,
	)
	webappPageChain := httproute.Chain(
		webappChain,
		p.Middleware(newCSRFMiddleware),
		webapp.TurbolinksMiddleware{},
		p.Middleware(newSecHeadersMiddleware),
	)
	webappAuthEntrypointChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newAuthEntryPointMiddleware),
		// A unique visit is started when the user visit auth entry point
		p.Middleware(newWebAppVisitorIDMiddleware),
	)
	webappAuthenticatedChain := httproute.Chain(
		webappPageChain,
		webapp.RequireAuthenticatedMiddleware{},
	)

	staticRoute := httproute.Route{Middleware: staticChain}
	oauthStaticRoute := httproute.Route{Middleware: oauthStaticChain}
	oauthAPIRoute := httproute.Route{Middleware: oauthAPIChain}
	apiRoute := httproute.Route{Middleware: apiChain}
	scopedRoute := httproute.Route{Middleware: scopedChain}
	webappPageRoute := httproute.Route{Middleware: webappPageChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}
	webappWebsocketRoute := httproute.Route{Middleware: webappWebsocketChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureOAuthEntrypointRoute(webappAuthEntrypointRoute), p.Handler(newWebAppOAuthEntrypointHandler))
	router.Add(webapphandler.ConfigureLoginRoute(webappAuthEntrypointRoute), p.Handler(newWebAppLoginHandler))
	router.Add(webapphandler.ConfigureSignupRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSignupHandler))
	router.Add(webapphandler.ConfigureSelectAccountRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSelectAccountHandler))

	router.Add(webapphandler.ConfigurePromoteRoute(webappPageRoute), p.Handler(newWebAppPromoteHandler))
	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappPageRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureSetupTOTPRoute(webappPageRoute), p.Handler(newWebAppSetupTOTPHandler))
	router.Add(webapphandler.ConfigureEnterTOTPRoute(webappPageRoute), p.Handler(newWebAppEnterTOTPHandler))
	router.Add(webapphandler.ConfigureSetupOOBOTPRoute(webappPageRoute), p.Handler(newWebAppSetupOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterOOBOTPRoute(webappPageRoute), p.Handler(newWebAppEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppEnterRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSetupRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppSetupRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureVerifyIdentityRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentityHandler))
	router.Add(webapphandler.ConfigureVerifyIdentitySuccessRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentitySuccessHandler))
	router.Add(webapphandler.ConfigureCreatePasswordRoute(webappPageRoute), p.Handler(newWebAppCreatePasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappPageRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppResetPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureUserDisabledRoute(webappPageRoute), p.Handler(newWebAppUserDisabledHandler))
	router.Add(webapphandler.ConfigureReturnRoute(webappPageRoute), p.Handler(newWebAppReturnHandler))
	router.Add(webapphandler.ConfigureErrorRoute(webappPageRoute), p.Handler(newWebAppErrorHandler))
	router.Add(webapphandler.ConfigureForceChangePasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangePasswordHandler))
	router.Add(webapphandler.ConfigureForceChangeSecondaryPasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangeSecondaryPasswordHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappAuthenticatedRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsBiometricRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsBiometricHandler))
	router.Add(webapphandler.ConfigureSettingsMFARoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsMFAHandler))
	router.Add(webapphandler.ConfigureSettingsTOTPRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsTOTPHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsRecoveryCodeRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSettingsSessionsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsSessionsHandler))
	router.Add(webapphandler.ConfigureSettingsChangePasswordRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsChangePasswordHandler))
	router.Add(webapphandler.ConfigureSettingsChangeSecondaryPasswordRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsChangeSecondaryPasswordHandler))

	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppSSOCallbackHandler))

	router.Add(webapphandler.ConfigureWechatAuthRoute(webappPageRoute), p.Handler(newWechatAuthHandler))
	router.Add(webapphandler.ConfigureWechatCallbackRoute(webappSSOCallbackRoute), p.Handler(newWechatCallbackHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(oauthStaticRoute), p.Handler(newOAuthJWKSHandler))

	router.Add(oauthhandler.ConfigureAuthorizeRoute(oauthAPIRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureFromWebAppRoute(oauthAPIRoute), p.Handler(newOAuthFromWebAppHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(oauthAPIRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(oauthAPIRoute), p.Handler(newOAuthRevokeHandler))
	router.Add(oauthhandler.ConfigureEndSessionRoute(oauthAPIRoute), p.Handler(newOAuthEndSessionHandler))

	router.Add(oauthhandler.ConfigureChallengeRoute(apiRoute), p.Handler(newOAuthChallengeHandler))
	router.Add(oauthhandler.ConfigureAppSessionTokenRoute(apiRoute), p.Handler(newOAuthAppSessionTokenHandler))

	router.Add(oauthhandler.ConfigureUserInfoRoute(scopedRoute), p.Handler(newOAuthUserInfoHandler))
	router.Add(webapphandler.ConfigureWebsocketRoute(webappWebsocketRoute), p.Handler(newWebAppWebsocketHandler))

	router.Add(webapphandler.ConfigureStaticAssetsRoute(staticRoute), p.Handler(newWebAppStaticAssetsHandler))

	return router
}
