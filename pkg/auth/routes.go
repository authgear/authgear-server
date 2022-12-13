package auth

import (
	apihandler "github.com/authgear/authgear-server/pkg/auth/handler/api"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	siwehandler "github.com/authgear/authgear-server/pkg/auth/handler/siwe"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func newUnsafeDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, true)
}

func newSafeDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, false)
}

func newAllSessionMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(deps, false)
}

func newIDPSessionOnlySessionMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(deps, true)
}

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {

	newSessionMiddleware := func(idpSessionOnly bool) httproute.Middleware {
		if idpSessionOnly {
			return p.Middleware(newIDPSessionOnlySessionMiddleware)
		}
		return p.Middleware(newAllSessionMiddleware)
	}

	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		httproute.MiddlewareFunc(httputil.StaticSecurityHeaders),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
	)

	// This route is intentionally simple.
	// This does not check Host and allow any origin.
	generatedStaticChain := httproute.Chain(
		httproute.MiddlewareFunc(httputil.StaticSecurityHeaders),
		httproute.MiddlewareFunc(middleware.CORSStar),
	)

	appStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
	)

	oauthStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
	)

	newOAuthAPIChain := func(idpSessionOnly bool) httproute.Middleware {
		return httproute.Chain(
			rootChain,
			p.Middleware(newCORSMiddleware),
			p.Middleware(newPublicOriginMiddleware),
			newSessionMiddleware(idpSessionOnly),
			httproute.MiddlewareFunc(httputil.NoStore),
			p.Middleware(newWebAppWeChatRedirectURIMiddleware),
		)
	}

	oauthAPIChain := newOAuthAPIChain(false)
	// authz endpoint only accepts idp session
	oauthAuthzAPIChain := newOAuthAPIChain(true)
	siweAPIChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		httproute.MiddlewareFunc(httputil.NoCache),
	)

	apiChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		p.Middleware(newAllSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoStore),
	)

	apiAuthenticatedChain := httproute.Chain(
		apiChain,
		p.Middleware(newAPIRRequireAuthenticatedMiddlewareMiddleware),
	)

	scopedChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		p.Middleware(newAllSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoStore),
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	newWebappChain := func(idpSessionOnly bool) httproute.Middleware {
		return httproute.Chain(
			rootChain,
			p.Middleware(newPublicOriginMiddleware),
			p.Middleware(newPanicWebAppMiddleware),
			newSessionMiddleware(idpSessionOnly),
			httproute.MiddlewareFunc(httputil.NoStore),
			httproute.MiddlewareFunc(webapp.IntlMiddleware),
			p.Middleware(newWebAppSessionMiddleware),
			p.Middleware(newWebAppUILocalesMiddleware),
			p.Middleware(newWebAppColorSchemeMiddleware),
			p.Middleware(newWebAppWeChatRedirectURIMiddleware),
			p.Middleware(newWebAppClientIDMiddleware),
			p.Middleware(newTutorialMiddleware),
		)
	}
	webappChain := newWebappChain(false)
	webappSSOCallbackChain := httproute.Chain(
		webappChain,
	)
	webappWebsocketChain := httproute.Chain(
		webappChain,
	)
	webappWhatsappCallbackChain := httproute.Chain(
		webappChain,
	)
	webappAPIChain := httproute.Chain(
		webappChain,
	)

	newWebappPageChain := func(idpSessionOnly bool) httproute.Middleware {
		return httproute.Chain(
			newWebappChain(idpSessionOnly),
			p.Middleware(newCSRFMiddleware),
			// Turbo no longer requires us to tell the redirected location.
			// It can now determine redirection from the response.
			// https://github.com/hotwired/turbo/blob/daabebb0575fffbae1b2582dc458967cd638e899/src/core/drive/visit.ts#L316
			p.Middleware(newSafeDynamicCSPMiddleware),
		)
	}
	webappPageChain := newWebappPageChain(false)
	webappSIWEChain := httproute.Chain(
		webappChain,
		p.Middleware(newCSRFMiddleware),
		p.Middleware(newUnsafeDynamicCSPMiddleware),
	)
	webappAuthEntrypointChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newAuthEntryPointMiddleware),
		// A unique visit is started when the user visit auth entry point
		p.Middleware(newWebAppVisitorIDMiddleware),
	)
	// select account page only accepts idp session
	webappSelectAccountChain := httproute.Chain(
		newWebappPageChain(true),
		p.Middleware(newAuthEntryPointMiddleware),
	)
	// consent page only accepts idp session
	webappConsentPageChain := newWebappPageChain(true)
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

	appStaticRoute := httproute.Route{Middleware: appStaticChain}
	generatedStaticRoute := httproute.Route{Middleware: generatedStaticChain}
	oauthStaticRoute := httproute.Route{Middleware: oauthStaticChain}
	oauthAPIRoute := httproute.Route{Middleware: oauthAPIChain}
	oauthAuthzAPIRoute := httproute.Route{Middleware: oauthAuthzAPIChain}
	siweAPIRoute := httproute.Route{Middleware: siweAPIChain}
	apiRoute := httproute.Route{Middleware: apiChain}
	apiAuthenticatedRoute := httproute.Route{Middleware: apiAuthenticatedChain}
	scopedRoute := httproute.Route{Middleware: scopedChain}
	webappPageRoute := httproute.Route{Middleware: webappPageChain}
	webappSIWERoute := httproute.Route{Middleware: webappSIWEChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappSelectAccountRoute := httproute.Route{Middleware: webappSelectAccountChain}
	webappConsentPageRoute := httproute.Route{Middleware: webappConsentPageChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSuccessPageRoute := httproute.Route{Middleware: webappSuccessPageChain}
	webappSettingsSubRoutesRoute := httproute.Route{Middleware: webappSettingsSubRoutesChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}
	webappWebsocketRoute := httproute.Route{Middleware: webappWebsocketChain}
	webappWhatsappCallbackRoute := httproute.Route{Middleware: webappWhatsappCallbackChain}
	webappAPIRoute := httproute.Route{Middleware: webappAPIChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureOAuthEntrypointRoute(webappAuthEntrypointRoute), p.Handler(newWebAppOAuthEntrypointHandler))
	router.Add(webapphandler.ConfigureLoginRoute(webappAuthEntrypointRoute), p.Handler(newWebAppLoginHandler))
	router.Add(webapphandler.ConfigureSignupRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSignupHandler))
	router.Add(webapphandler.ConfigureSelectAccountRoute(webappSelectAccountRoute), p.Handler(newWebAppSelectAccountHandler))

	router.Add(webapphandler.ConfigurePromoteRoute(webappPageRoute), p.Handler(newWebAppPromoteHandler))
	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappPageRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureUsePasskeyRoute(webappPageRoute), p.Handler(newWebAppUsePasskeyHandler))
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
	router.Add(webapphandler.ConfigureCreatePasskeyRoute(webappPageRoute), p.Handler(newWebAppCreatePasskeyHandler))
	router.Add(webapphandler.ConfigurePromptCreatePasskeyRoute(webappPageRoute), p.Handler(newWebAppPromptCreatePasskeyHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappPageRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureAccountStatusRoute(webappPageRoute), p.Handler(newWebAppAccountStatusHandler))
	router.Add(webapphandler.ConfigureReturnRoute(webappPageRoute), p.Handler(newWebAppReturnHandler))
	router.Add(webapphandler.ConfigureErrorRoute(webappPageRoute), p.Handler(newWebAppErrorHandler))
	router.Add(webapphandler.ConfigureForceChangePasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangePasswordHandler))
	router.Add(webapphandler.ConfigureForceChangeSecondaryPasswordRoute(webappPageRoute), p.Handler(newWebAppForceChangeSecondaryPasswordHandler))
	router.Add(webapphandler.ConfigureConnectWeb3AccountRoute(webappSIWERoute), p.Handler(newWebAppConnectWeb3AccountHandler))
	router.Add(webapphandler.ConfigureMissingWeb3WalletRoute(webappPageRoute), p.Handler(newWebAppMissingWeb3WalletHandler))

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
	router.Add(webapphandler.ConfigureSettingsPasskeyRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsPasskeyHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsRecoveryCodeRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSettingsSessionsRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsSessionsHandler))
	router.Add(webapphandler.ConfigureSettingsChangePasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsChangePasswordHandler))
	router.Add(webapphandler.ConfigureSettingsChangeSecondaryPasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsChangeSecondaryPasswordHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsDeleteAccountHandler))

	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppSSOCallbackHandler))

	router.Add(webapphandler.ConfigureWechatAuthRoute(webappPageRoute), p.Handler(newWechatAuthHandler))
	router.Add(webapphandler.ConfigureWechatCallbackRoute(webappSSOCallbackRoute), p.Handler(newWechatCallbackHandler))

	router.Add(webapphandler.ConfigureWhatsappWATICallbackRoute(webappWhatsappCallbackRoute), p.Handler(newWhatsappWATICallbackHandler))

	router.Add(webapphandler.ConfigurePasskeyCreationOptionsRoute(webappAPIRoute), p.Handler(newWebAppPasskeyCreationOptionsHandler))
	router.Add(webapphandler.ConfigurePasskeyRequestOptionsRoute(webappAPIRoute), p.Handler(newWebAppPasskeyRequestOptionsHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(oauthStaticRoute), p.Handler(newOAuthJWKSHandler))

	router.Add(oauthhandler.ConfigureAuthorizeRoute(oauthAuthzAPIRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(oauthAPIRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(oauthAPIRoute), p.Handler(newOAuthRevokeHandler))
	router.Add(oauthhandler.ConfigureEndSessionRoute(oauthAPIRoute), p.Handler(newOAuthEndSessionHandler))

	router.Add(oauthhandler.ConfigureChallengeRoute(apiRoute), p.Handler(newOAuthChallengeHandler))
	router.Add(oauthhandler.ConfigureAppSessionTokenRoute(apiRoute), p.Handler(newOAuthAppSessionTokenHandler))

	router.Add(oauthhandler.ConfigureUserInfoRoute(scopedRoute), p.Handler(newOAuthUserInfoHandler))

	router.Add(oauthhandler.ConfigureConsentRoute(webappConsentPageRoute), p.Handler(newOAuthConsentHandler))

	router.Add(siwehandler.ConfigureNonceRoute(siweAPIRoute), p.Handler(newSIWENonceHandler))

	router.Add(apihandler.ConfigureAnonymousUserSignupRoute(apiRoute), p.Handler(newAPIAnonymousUserSignupHandler))
	router.Add(apihandler.ConfigureAnonymousUserPromotionCodeRoute(apiRoute), p.Handler(newAPIAnonymousUserPromotionCodeHandler))
	router.Add(apihandler.ConfigurePresignImagesUploadRoute(apiAuthenticatedRoute), p.Handler(newAPIPresignImagesUploadHandler))

	router.Add(webapphandler.ConfigureWebsocketRoute(webappWebsocketRoute), p.Handler(newWebAppWebsocketHandler))

	router.Add(webapphandler.ConfigureAppStaticAssetsRoute(appStaticRoute), p.Handler(newWebAppAppStaticAssetsHandler))

	router.Add(webapphandler.ConfigureGeneratedStaticAssetsRoute(generatedStaticRoute), p.RootHandler(newWebAppGeneratedStaticAssetsHandler))

	router.NotFound(webappPageRoute, p.Handler(newWebAppNotFoundHandler))

	return router
}
