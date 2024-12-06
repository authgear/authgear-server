package auth

import (
	"net/http"

	apihandler "github.com/authgear/authgear-server/pkg/auth/handler/api"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	samlhandler "github.com/authgear/authgear-server/pkg/auth/handler/saml"
	siwehandler "github.com/authgear/authgear-server/pkg/auth/handler/siwe"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	webapphandlerauthflowv2 "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httproute/httprouteotel"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func newSIWEDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowFrameAncestorsFromEnv(true), webapp.AllowFrameAncestorsFromCustomUI(false))
}

func newWebPageDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowFrameAncestorsFromEnv(true), webapp.AllowFrameAncestorsFromCustomUI(false))
}

func newConsentPageDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowFrameAncestorsFromEnv(false), webapp.AllowFrameAncestorsFromCustomUI(true))
}

func newAllSessionMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(deps)
}

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) http.Handler {

	newSessionMiddleware := func() httproute.Middleware {
		return p.Middleware(newAllSessionMiddleware)
	}

	router := httprouteotel.NewOTelRouter(httproute.NewRouter())

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		RequestMiddleware(p, configSource, newRequestMiddleware),
		p.Middleware(newContextHolderMiddleware),
	)

	// This route is intentionally simple.
	// This does not check Host and allow any origin.
	generatedStaticChain := httproute.Chain(
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		httproute.MiddlewareFunc(middleware.CORSStar),
		httputil.GzipMiddleware{},
	)

	appStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		httputil.GzipMiddleware{},
	)

	samlStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
	)

	samlAPIChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		newSessionMiddleware(),
		httproute.MiddlewareFunc(httputil.NoStore),
		p.Middleware(newWebAppWeChatRedirectURIMiddleware),
	)

	oauthStaticChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
	)

	newOAuthAPIChain := func() httproute.Middleware {
		return httproute.Chain(
			rootChain,
			p.Middleware(newCORSMiddleware),
			p.Middleware(newPublicOriginMiddleware),
			newSessionMiddleware(),
			httproute.MiddlewareFunc(httputil.NoStore),
			p.Middleware(newWebAppWeChatRedirectURIMiddleware),
		)
	}

	oauthAPIChain := newOAuthAPIChain()
	dpopOAuthAPIChain := httproute.Chain(
		oauthAPIChain,
		p.Middleware(newDPoPMiddleware),
	)
	oauthAuthzAPIChain := newOAuthAPIChain()
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
	dpopApiChain := httproute.Chain(
		apiChain,
		p.Middleware(newDPoPMiddleware),
	)

	workflowChain := httproute.Chain(
		apiChain,
		p.Middleware(newWorkflowIntlMiddleware),
	)

	authenticationFlowChain := httproute.Chain(
		apiChain,
		p.Middleware(newAuthenticationFlowIntlMiddleware),
		p.Middleware(newAuthenticationFlowRateLimitMiddleware),
	)

	apiAuthenticatedChain := httproute.Chain(
		apiChain,
		p.Middleware(newAPIRRequireAuthenticatedMiddlewareMiddleware),
	)

	accountManagementChain := httproute.Chain(
		apiChain,
		p.Middleware(newAccountManagementRateLimitMiddleware),
		p.Middleware(newAPIRRequireAuthenticatedMiddlewareMiddleware),
	)

	oauthAPIScopedChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
		p.Middleware(newPublicOriginMiddleware),
		p.Middleware(newAllSessionMiddleware),
		httproute.MiddlewareFunc(httputil.NoStore),
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	newWebappChain := func() httproute.Middleware {
		return httproute.Chain(
			rootChain,
			p.Middleware(newPublicOriginMiddleware),
			p.Middleware(newPanicWebAppMiddleware),
			newSessionMiddleware(),
			httproute.MiddlewareFunc(httputil.NoStore),
			httproute.MiddlewareFunc(webapp.IntlMiddleware),
			p.Middleware(newWebAppSessionMiddleware),
			p.Middleware(newWebAppUIParamMiddleware),
			p.Middleware(newWebAppColorSchemeMiddleware),
			p.Middleware(newWebAppWeChatRedirectURIMiddleware),
			p.Middleware(newTutorialMiddleware),
		)
	}
	webappChain := newWebappChain()
	webappSSOCallbackChain := httproute.Chain(
		webappChain,
	)
	webappWebsocketChain := httproute.Chain(
		webappChain,
	)
	webappAPIChain := httproute.Chain(
		webappChain,
	)

	webappNotFoundChain := httproute.Chain(
		newWebappChain(),
		p.Middleware(newWebPageDynamicCSPMiddleware),
	)

	newWebappPageChain := func() httproute.Middleware {
		return httproute.Chain(
			newWebappChain(),
			p.Middleware(newCSRFDebugMiddleware),
			p.Middleware(newCSRFMiddleware),
			// Turbo no longer requires us to tell the redirected location.
			// It can now determine redirection from the response.
			// https://github.com/hotwired/turbo/blob/daabebb0575fffbae1b2582dc458967cd638e899/src/core/drive/visit.ts#L316
			p.Middleware(newWebPageDynamicCSPMiddleware),
			p.Middleware(newWebAppWeChatRedirectURIMiddleware),
		)
	}
	webappPageChain := newWebappPageChain()
	webappAuthEntrypointChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newAuthEntryPointMiddleware),
		// A unique visit is started when the user visit auth entry point
		p.Middleware(newWebAppVisitorIDMiddleware),
	)
	webappRequireAuthEnabledAuthEntrypointChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newRequireAuthenticationEnabledMiddleware),
		p.Middleware(newAuthEntryPointMiddleware),
		// A unique visit is started when the user visit auth entry point
		p.Middleware(newWebAppVisitorIDMiddleware),
	)
	webappPromoteChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newRequireAuthenticationEnabledMiddleware),
		p.Middleware(newAuthEntryPointMiddleware),
	)
	webappReauthChain := httproute.Chain(
		newWebappPageChain(),
		p.Middleware(newAuthEntryPointMiddleware),
	)
	webappSelectAccountChain := httproute.Chain(
		newWebappPageChain(),
	)
	webappVerifyBotProtectionChain := httproute.Chain(
		webappPageChain,
	)
	// consent page only accepts idp session
	webappConsentPageChain := httproute.Chain(
		newWebappChain(),
		p.Middleware(newCSRFDebugMiddleware),
		p.Middleware(newCSRFMiddleware),
		p.Middleware(newConsentPageDynamicCSPMiddleware),
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
	webappSettingsChain := httproute.Chain(
		webappAuthenticatedChain,
		p.Middleware(newRequireSettingsEnabledMiddleware),
	)
	webappSettingsSubRoutesChain := httproute.Chain(
		webappAuthenticatedChain,
		p.Middleware(newWebAppSessionMiddleware),
		p.Middleware(newRequireSettingsEnabledMiddleware),
		// SettingsSubRoutesMiddleware should be added to all the settings sub routes only
		// but no /settings itself
		// it redirects all sub routes to /settings if the current user is
		// anonymous user
		p.Middleware(newSettingsSubRoutesMiddleware),
	)
	webappPagePreviewChain := httproute.Chain(
		rootChain,
		p.Middleware(newPublicOriginMiddleware),
		p.Middleware(newPanicWebAppMiddleware),
		httproute.MiddlewareFunc(httputil.NoStore),
		httproute.MiddlewareFunc(webapp.IntlMiddleware),
		p.Middleware(newWebAppColorSchemeMiddleware),
		p.Middleware(newWebPageDynamicCSPMiddleware),
	)

	appStaticRoute := httproute.Route{Middleware: appStaticChain}
	generatedStaticRoute := httproute.Route{Middleware: generatedStaticChain}
	samlStaticRoute := httproute.Route{Middleware: samlStaticChain}
	samlAPIRoute := httproute.Route{Middleware: samlAPIChain}
	oauthStaticRoute := httproute.Route{Middleware: oauthStaticChain}
	oauthAPIRoute := httproute.Route{Middleware: oauthAPIChain}
	dpopOauthAPIRoute := httproute.Route{Middleware: dpopOAuthAPIChain}
	oauthAuthzAPIRoute := httproute.Route{Middleware: oauthAuthzAPIChain}
	siweAPIRoute := httproute.Route{Middleware: siweAPIChain}
	apiRoute := httproute.Route{Middleware: apiChain}
	dpopApiRoute := httproute.Route{Middleware: dpopApiChain}
	workflowRoute := httproute.Route{Middleware: workflowChain}
	authenticationFlowRoute := httproute.Route{Middleware: authenticationFlowChain}
	apiAuthenticatedRoute := httproute.Route{Middleware: apiAuthenticatedChain}
	accountManagementRoute := httproute.Route{Middleware: accountManagementChain}
	oauthAPIScopedRoute := httproute.Route{Middleware: oauthAPIScopedChain}
	webappPageRoute := httproute.Route{Middleware: webappPageChain}
	webappNotFoundRoute := httproute.Route{Middleware: webappNotFoundChain}
	webappPromoteRoute := httproute.Route{Middleware: webappPromoteChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappRequireAuthEnabledAuthEntrypointRoute := httproute.Route{Middleware: webappRequireAuthEnabledAuthEntrypointChain}
	webappSelectAccountRoute := httproute.Route{Middleware: webappSelectAccountChain}
	webappReauthRoute := httproute.Route{Middleware: webappReauthChain}
	webappVerifyBotProtectionRoute := httproute.Route{Middleware: webappVerifyBotProtectionChain}
	webappConsentPageRoute := httproute.Route{Middleware: webappConsentPageChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSuccessPageRoute := httproute.Route{Middleware: webappSuccessPageChain}
	webappSettingsRoute := httproute.Route{Middleware: webappSettingsChain}
	webappSettingsSubRoutesRoute := httproute.Route{Middleware: webappSettingsSubRoutesChain}
	webappTesterRouter := httproute.Route{Middleware: webappChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}
	webappWebsocketRoute := httproute.Route{Middleware: webappWebsocketChain}
	webappAPIRoute := httproute.Route{Middleware: webappAPIChain}
	webappPagePreviewRoute := httproute.Route{Middleware: webappPagePreviewChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureOAuthEntrypointRoute(webappAuthEntrypointRoute), p.Handler(newWebAppOAuthEntrypointHandler))
	router.Add(webapphandler.ConfigureAuthflowLoginRoute(webappRequireAuthEnabledAuthEntrypointRoute), p.Handler(newWebAppAuthflowV2LoginHandler))
	router.Add(webapphandler.ConfigureAuthflowSignupRoute(webappRequireAuthEnabledAuthEntrypointRoute), p.Handler(newWebAppAuthflowV2SignupHandler))
	router.Add(webapphandler.ConfigureAuthflowPromoteRoute(webappPromoteRoute), p.Handler(newWebAppAuthflowV2PromoteHandler))
	router.Add(webapphandler.ConfigureAuthflowReauthRoute(webappReauthRoute), p.Handler(newWebAppAuthflowV2ReauthHandler))
	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppAuthflowV2SSOCallbackHandler))
	router.Add(webapphandler.ConfigureWechatCallbackRoute(webappSSOCallbackRoute), p.Handler(newWechatCallbackHandler))

	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SelectAccountRoute(webappSelectAccountRoute), p.Handler(newWebAppAuthflowV2SelectAccountHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2VerifyBotProtectionRoute(webappVerifyBotProtectionRoute), p.Handler(newWebAppAuthflowV2VerifyBotProtectionHandler))

	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2EnterPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2EnterPasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2LDAPLoginRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2LDAPLoginHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2EnterOOBOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2EnterOOBOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SetupOOBOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2SetupOOBOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ViewRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ViewRecoveryCodeHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ErrorRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ErrorHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2CreatePasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2CreatePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2AccountStatusRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2AccountStatusHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2EnterRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2EnterRecoveryCodeHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowv2ChangePasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ChangePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ChangePasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ChangePasswordSuccessHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ForgotPasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ForgotPasswordLinkSentRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ForgotPasswordLinkSentHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ForgotPasswordOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ForgotPasswordOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ResetPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ResetPasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2ResetPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2ResetPasswordSuccessHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SetupTOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2SetupTOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2EnterTOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2EnterTOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureV2AuthflowOOBOTPLinkRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2OOBOTPLinkHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2VerifyLoginLinkOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2VerifyLoginLinkOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PromptCreatePasskeyRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2PromptCreatePasskeyHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2UsePasskeyRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2UsePasskeyHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2TerminateOtherSessionsRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2TerminateOtherSessionsHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2NoAuthenticatorRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2NoAuthenticatorHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowv2FinishFlowRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2FinishFlowHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2WechatRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2WechatHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2AccountLinkingRoute(webappPageRoute), p.Handler(newWebAppAuthflowV2AccountLinkingHandler))

	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandler.ConfigureAuthflowLoginRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2LoginHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandler.ConfigureAuthflowSignupRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2SignupHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2EnterPasswordRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2EnterPasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2EnterOOBOTPRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2EnterOOBOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2UsePasskeyRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2UsePasskeyHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2PromptCreatePasskeyRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2PromptCreatePasskeyHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2EnterTOTPRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2EnterTOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureV2AuthflowOOBOTPLinkRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2OOBOTPLinkHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2CreatePasswordRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2CreatePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2PreviewRoute(webapphandlerauthflowv2.ConfigureAuthflowV2ErrorRoute(webappPagePreviewRoute)), p.Handler(newWebAppAuthflowV2ErrorHandler))

	router.Add(webapphandler.ConfigureReturnRoute(webappPageRoute), p.Handler(newWebAppReturnHandler))
	router.Add(webapphandler.ConfigureErrorRoute(webappPageRoute), p.Handler(newWebAppErrorHandler))
	router.Add(webapphandler.ConfigureFeatureDisabledRoute(webappPageRoute), p.Handler(newWebAppFeatureDisabledHandler))

	router.Add(webapphandler.ConfigureSettingsDeleteAccountSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppAuthflowV2SettingsDeleteAccountSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappSettingsRoute), p.Handler(newWebAppAuthflowV2SettingsHandler))

	router.Add(webapphandler.ConfigureSettingsProfileRoute(webappSettingsRoute), p.Handler(newWebAppAuthflowV2SettingsProfile))
	router.Add(webapphandler.ConfigureSettingsProfileEditRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsProfileEditHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityChangePrimaryEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityChangePrimaryEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityAddEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityAddEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityEditEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityEditEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityVerifyEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityVerifyEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityViewEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityViewEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityAddPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityAddPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityChangePrimaryPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityChangePrimaryPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityEditPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityEditPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityViewPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityViewPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityVerifyPhoneRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityVerifyPhoneHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListUsername(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListUsernameHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityNewUsername(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityNewUsernameHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityViewUsername(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityViewUsernameHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityEditUsername(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityEditUsernameHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListOAuthRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListOAuthHandler))
	router.Add(webapphandler.ConfigureSettingsBiometricRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsBiometricHandler))
	router.Add(webapphandler.ConfigureSettingsMFARoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAViewRecoveryCodeRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAViewRecoveryCodeHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFACreatePassword(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFACreatePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAChangePassword(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAChangePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAPassword(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAPasswordHandler))
	router.Add(webapphandler.ConfigureSettingsTOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsTOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFACreateTOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFACreateTOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAEnterTOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAEnterTOTPHandler))
	router.Add(webapphandler.ConfigureSettingsPasskeyRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsChangePasskeyHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsOOBOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFACreateOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFACreateOOBOTPHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAEnterOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsSessionsRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsSessionsHandler))
	router.Add(webapphandler.ConfigureSettingsChangePasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsChangePasswordHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsDeleteAccountHandler))
	router.Add(webapphandlerauthflowv2.ConfigureSettingsV2AdvancedSettingsRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsAdvancedSettingsHandler))

	router.Add(webapphandler.ConfigureTesterRoute(webappTesterRouter), p.Handler(newWebAppTesterHandler))

	router.Add(webapphandler.ConfigurePasskeyCreationOptionsRoute(webappAPIRoute), p.Handler(newWebAppPasskeyCreationOptionsHandler))
	router.Add(webapphandler.ConfigurePasskeyRequestOptionsRoute(webappAPIRoute), p.Handler(newWebAppPasskeyRequestOptionsHandler))

	router.Add(webapphandler.ConfigureCSRFErrorInstructionRoute(webappAPIRoute), p.Handler(newWebAppCSRFErrorInstructionHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(oauthStaticRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(oauthStaticRoute), p.Handler(newOAuthJWKSHandler))

	router.Add(oauthhandler.ConfigureAuthorizeRoute(oauthAuthzAPIRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(dpopOauthAPIRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(dpopOauthAPIRoute), p.Handler(newOAuthRevokeHandler))
	router.Add(oauthhandler.ConfigureEndSessionRoute(oauthAPIRoute), p.Handler(newOAuthEndSessionHandler))

	router.Add(oauthhandler.ConfigureChallengeRoute(apiRoute), p.Handler(newOAuthChallengeHandler))
	router.Add(oauthhandler.ConfigureAppSessionTokenRoute(dpopApiRoute), p.Handler(newOAuthAppSessionTokenHandler))
	router.Add(oauthhandler.ConfigureProxyRedirectRoute(apiRoute), p.Handler(newOAuthProxyRedirectHandler))

	router.Add(oauthhandler.ConfigureUserInfoRoute(oauthAPIScopedRoute), p.Handler(newOAuthUserInfoHandler))

	router.Add(oauthhandler.ConfigureConsentRoute(webappConsentPageRoute), p.Handler(newOAuthConsentHandler))

	router.Add(samlhandler.ConfigureMetadataRoute(samlStaticRoute), p.Handler(newSAMLMetadataHandler))
	router.Add(samlhandler.ConfigureLoginRoute(samlAPIRoute), p.Handler(newSAMLLoginHandler))
	router.Add(samlhandler.ConfigureLoginFinishRoute(samlAPIRoute), p.Handler(newSAMLLoginFinishHandler))
	router.Add(samlhandler.ConfigureLogoutRoute(samlAPIRoute), p.Handler(newSAMLLogoutHandler))

	router.Add(siwehandler.ConfigureNonceRoute(siweAPIRoute), p.Handler(newSIWENonceHandler))

	router.Add(apihandler.ConfigureAnonymousUserSignupRoute(apiRoute), p.Handler(newAPIAnonymousUserSignupHandler))
	router.Add(apihandler.ConfigureAnonymousUserPromotionCodeRoute(apiRoute), p.Handler(newAPIAnonymousUserPromotionCodeHandler))
	router.Add(apihandler.ConfigurePresignImagesUploadRoute(apiAuthenticatedRoute), p.Handler(newAPIPresignImagesUploadHandler))

	router.Add(webapphandler.ConfigureWebsocketRoute(webappWebsocketRoute), p.Handler(newWebAppWebsocketHandler))

	router.Add(webapphandler.ConfigureAppStaticAssetsRoute(appStaticRoute), p.Handler(newWebAppAppStaticAssetsHandler))

	router.Add(webapphandler.ConfigureGeneratedStaticAssetsRoute(generatedStaticRoute), p.RootHandler(newWebAppGeneratedStaticAssetsHandler))

	router.Add(apihandler.ConfigureWorkflowNewRoute(workflowRoute), p.Handler(newAPIWorkflowNewHandler))
	router.Add(apihandler.ConfigureWorkflowGetRoute(workflowRoute), p.Handler(newAPIWorkflowGetHandler))
	router.Add(apihandler.ConfigureWorkflowInputRoute(workflowRoute), p.Handler(newAPIWorkflowInputHandler))
	router.Add(apihandler.ConfigureWorkflowWebsocketRoute(workflowRoute), p.Handler(newAPIWorkflowWebsocketHandler))
	router.Add(apihandler.ConfigureWorkflowV2Route(workflowRoute), p.Handler(newAPIWorkflowV2Handler))

	router.Add(apihandler.ConfigureAuthenticationFlowV1CreateRoute(authenticationFlowRoute), p.Handler(newAPIAuthenticationFlowV1CreateHandler))
	router.Add(apihandler.ConfigureAuthenticationFlowV1InputRoute(authenticationFlowRoute), p.Handler(newAPIAuthenticationFlowV1InputHandler))
	router.Add(apihandler.ConfigureAuthenticationFlowV1GetRoute(authenticationFlowRoute), p.Handler(newAPIAuthenticationFlowV1GetHandler))
	router.Add(apihandler.ConfigureAuthenticationFlowV1WebsocketRoute(authenticationFlowRoute), p.Handler(newAPIAuthenticationFlowV1WebsocketHandler))

	router.Add(apihandler.ConfigureAccountManagementV1IdentificationRoute(accountManagementRoute), p.Handler(newAPIAccountManagementV1IdentificationHandler))
	router.Add(apihandler.ConfigureAccountManagementV1IdentificationOAuthRoute(accountManagementRoute), p.Handler(newAPIAccountManagementV1IdentificationOAuthHandler))

	router.NotFound(webappNotFoundRoute, p.Handler(newWebAppAuthflowV2NotFoundHandler))

	return router.HTTPHandler()
}
