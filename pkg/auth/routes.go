package auth

import (
	apihandler "github.com/authgear/authgear-server/pkg/auth/handler/api"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/web"
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
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
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
		p.Middleware(newCORSMiddleware),
	)

	apiAuthenticatedChain := httproute.Chain(
		apiChain,
		p.Middleware(newAPIRRequireAuthenticatedMiddlewareMiddleware),
	)

	scopedChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		p.Middleware(newCORSMiddleware),
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	webappChain := httproute.Chain(
		rootChain,
		p.Middleware(newPanicWebAppMiddleware),
		p.Middleware(newSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
		httproute.MiddlewareFunc(webapp.IntlMiddleware),
		p.Middleware(newWebAppSessionMiddleware),
		p.Middleware(newWebAppUILocalesMiddleware),
		p.Middleware(newWebAppColorSchemeMiddleware),
		p.Middleware(newWebAppWeChatRedirectURIMiddleware),
		p.Middleware(newWebAppClientIDMiddleware),
		p.Middleware(newTutorialMiddleware),
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
		web.StaticSecurityHeadersMiddleware{},
		p.Middleware(newDynamicCSPMiddleware),
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
	webappSuccessPageChain := httproute.Chain(
		webappPageChain,
		// SuccessPageMiddleware check the cookie and see if it is valid to
		// visit the success page
		p.Middleware(newSuccessPageMiddleware),
	)
	webappSettingsSubRoutesChain := httproute.Chain(
		webappAuthenticatedChain,
		// SettingsSubRoutesMiddleware should be added to all the settings sub routes only
		// but no /settings itself
		// it redirects all sub routes to /settings if the current user is
		// anonymous user
		p.Middleware(newSettingsSubRoutesMiddleware),
	)

	staticRoute := httproute.Route{Middleware: staticChain}
	oauthStaticRoute := httproute.Route{Middleware: oauthStaticChain}
	oauthAPIRoute := httproute.Route{Middleware: oauthAPIChain}
	apiRoute := httproute.Route{Middleware: apiChain}
	apiAuthenticatedRoute := httproute.Route{Middleware: apiAuthenticatedChain}
	scopedRoute := httproute.Route{Middleware: scopedChain}
	webappPageRoute := httproute.Route{Middleware: webappPageChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSuccessPageRoute := httproute.Route{Middleware: webappSuccessPageChain}
	webappSettingsSubRoutesRoute := httproute.Route{Middleware: webappSettingsSubRoutesChain}
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
	router.Add(webapphandler.ConfigureSetupWhatsappOTPRoute(webappPageRoute), p.Handler(newWebAppSetupWhatsappOTPHandler))
	router.Add(webapphandler.ConfigureWhatsappOTPRoute(webappPageRoute), p.Handler(newWebAppWhatsappOTPHandler))
	router.Add(webapphandler.ConfigureEnterRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppEnterRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSetupRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppSetupRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureVerifyIdentityRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentityHandler))
	router.Add(webapphandler.ConfigureVerifyIdentitySuccessRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentitySuccessHandler))
	router.Add(webapphandler.ConfigureCreatePasswordRoute(webappPageRoute), p.Handler(newWebAppCreatePasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappPageRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureAccountStatusRoute(webappPageRoute), p.Handler(newWebAppAccountStatusHandler))
	router.Add(webapphandler.ConfigureReturnRoute(webappPageRoute), p.Handler(newWebAppReturnHandler))
	router.Add(webapphandler.ConfigureErrorRoute(webappPageRoute), p.Handler(newWebAppErrorHandler))
	router.Add(webapphandler.ConfigureForceChangePasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangePasswordHandler))
	router.Add(webapphandler.ConfigureForceChangeSecondaryPasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangeSecondaryPasswordHandler))

	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppResetPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppSettingsDeleteAccountSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappAuthenticatedRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsHandler))

	router.Add(webapphandler.ConfigureSettingsProfileRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsProfileHandler))
	router.Add(webapphandler.ConfigureSettingsProfileEditRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsProfileEditHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsBiometricRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsBiometricHandler))
	router.Add(webapphandler.ConfigureSettingsMFARoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsMFAHandler))
	router.Add(webapphandler.ConfigureSettingsTOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsTOTPHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsRecoveryCodeRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSettingsSessionsRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsSessionsHandler))
	router.Add(webapphandler.ConfigureSettingsChangePasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsChangePasswordHandler))
	router.Add(webapphandler.ConfigureSettingsChangeSecondaryPasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsChangeSecondaryPasswordHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsDeleteAccountHandler))

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

	router.Add(apihandler.ConfigureAnonymousUserSignupRoute(apiRoute), p.Handler(newAPIAnonymousUserSignupHandler))
	router.Add(apihandler.ConfigureAnonymousUserPromotionCodeRoute(apiRoute), p.Handler(newAPIAnonymousUserPromotionCodeHandler))
	router.Add(apihandler.ConfigurePresignImagesUploadRoute(apiAuthenticatedRoute), p.Handler(newAPIPresignImagesUploadHandler))

	router.Add(webapphandler.ConfigureWebsocketRoute(webappWebsocketRoute), p.Handler(newWebAppWebsocketHandler))

	router.Add(webapphandler.ConfigureStaticAssetsRoute(staticRoute), p.Handler(newWebAppStaticAssetsHandler))

	return router
}
