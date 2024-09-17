package auth

import (
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
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func newSIWEDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowInlineScript(true), webapp.AllowFrameAncestorsFromEnv(true), webapp.AllowFrameAncestorsFromCustomUI(false))
}

func newWebPageDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowInlineScript(false), webapp.AllowFrameAncestorsFromEnv(true), webapp.AllowFrameAncestorsFromCustomUI(false))
}

func newConsentPageDynamicCSPMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newDynamicCSPMiddleware(deps, webapp.AllowInlineScript(false), webapp.AllowFrameAncestorsFromEnv(false), webapp.AllowFrameAncestorsFromCustomUI(true))
}

func newAllSessionMiddleware(deps *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(deps)
}

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {

	newSessionMiddleware := func() httproute.Middleware {
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
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
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
			p.Middleware(newImplementationSwitcherMiddleware),
			p.Middleware(newSettingImplementationSwitcherMiddleware),
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
	webappSIWEChain := httproute.Chain(
		webappChain,
		p.Middleware(newCSRFDebugMiddleware),
		p.Middleware(newCSRFMiddleware),
		p.Middleware(newSIWEDynamicCSPMiddleware),
	)
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
	webappPromoteRoute := httproute.Route{Middleware: webappPromoteChain}
	webappSIWERoute := httproute.Route{Middleware: webappSIWEChain}
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
	router.Add(webapphandler.ConfigureAuthflowLoginRoute(webappRequireAuthEnabledAuthEntrypointRoute), &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppLoginHandler),
		Authflow:    p.Handler(newWebAppAuthflowLoginHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2LoginHandler),
	})
	router.Add(webapphandler.ConfigureAuthflowSignupRoute(webappRequireAuthEnabledAuthEntrypointRoute), &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppSignupHandler),
		Authflow:    p.Handler(newWebAppAuthflowSignupHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2SignupHandler),
	})
	router.Add(webapphandler.ConfigureAuthflowPromoteRoute(webappPromoteRoute), &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppPromoteHandler),
		Authflow:    p.Handler(newWebAppAuthflowPromoteHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2PromoteHandler),
	})
	router.Add(webapphandler.ConfigureAuthflowReauthRoute(webappReauthRoute), &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppReauthHandler),
		Authflow:    p.Handler(newWebAppAuthflowReauthHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2ReauthHandler),
	})
	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppSSOCallbackHandler),
		Authflow:    p.Handler(newWebAppAuthflowSSOCallbackHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2SSOCallbackHandler),
	})
	router.Add(webapphandler.ConfigureWechatCallbackRoute(webappSSOCallbackRoute), p.Handler(newWechatCallbackHandler))

	router.Add(webapphandler.ConfigureSelectAccountRoute(webappSelectAccountRoute), p.Handler(newWebAppSelectAccountHandler))
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

	router.Add(webapphandler.ConfigureAuthflowEnterPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowEnterPasswordHandler))
	router.Add(webapphandler.ConfigureAuthflowEnterOOBOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowCreatePasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowCreatePasswordHandler))
	router.Add(webapphandler.ConfigureAuthflowEnterTOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowEnterTOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowSetupTOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowSetupTOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowViewRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppAuthflowViewRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureAuthflowWhatsappOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowWhatsappOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowOOBOTPLinkRoute(webappPageRoute), p.Handler(newWebAppAuthflowOOBOTPLinkHandler))
	router.Add(webapphandler.ConfigureAuthflowChangePasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowChangePasswordHandler))
	router.Add(webapphandler.ConfigureAuthflowUsePasskeyRoute(webappPageRoute), p.Handler(newWebAppAuthflowUsePasskeyHandler))
	router.Add(webapphandler.ConfigureAuthflowPromptCreatePasskeyRoute(webappPageRoute), p.Handler(newWebAppAuthflowPromptCreatePasskeyHandler))
	router.Add(webapphandler.ConfigureAuthflowEnterRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppAuthflowEnterRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureAuthflowSetupOOBOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowSetupOOBOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowTerminateOtherSessionsRoute(webappPageRoute), p.Handler(newWebAppAuthflowTerminateOtherSessionsHandler))
	router.Add(webapphandler.ConfigureAuthflowWechatRoute(webappPageRoute), p.Handler(newWebAppAuthflowWechatHandler))
	router.Add(webapphandler.ConfigureAuthflowForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowForgotPasswordHandler))
	router.Add(webapphandler.ConfigureAuthflowForgotPasswordOTPRoute(webappPageRoute), p.Handler(newWebAppAuthflowForgotPasswordOTPHandler))
	router.Add(webapphandler.ConfigureAuthflowForgotPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppAuthflowForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureAuthflowResetPasswordRoute(webappPageRoute), p.Handler(newWebAppAuthflowResetPasswordHandler))
	router.Add(webapphandler.ConfigureAuthflowResetPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppAuthflowResetPasswordSuccessHandler))

	router.Add(webapphandler.ConfigureAuthflowAccountStatusRoute(webappPageRoute), p.Handler(newWebAppAuthflowAccountStatusHandler))
	router.Add(webapphandler.ConfigureAuthflowNoAuthenticatorRoute(webappPageRoute), p.Handler(newWebAppAuthflowNoAuthenticatorHandler))
	router.Add(webapphandler.ConfigureAuthflowFinishFlowRoute(webappPageRoute), p.Handler(newWebAppAuthflowFinishFlowHandler))

	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappPageRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureConfirmTerminateOtherSessionsRoute(webappPageRoute), p.Handler(newWebConfirmTerminateOtherSessionsHandler))
	router.Add(webapphandler.ConfigureUsePasskeyRoute(webappPageRoute), p.Handler(newWebAppUsePasskeyHandler))
	router.Add(webapphandler.ConfigureSetupTOTPRoute(webappPageRoute), p.Handler(newWebAppSetupTOTPHandler))
	router.Add(webapphandler.ConfigureEnterTOTPRoute(webappPageRoute), p.Handler(newWebAppEnterTOTPHandler))
	router.Add(webapphandler.ConfigureSetupOOBOTPRoute(webappPageRoute), p.Handler(newWebAppSetupOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterOOBOTPRoute(webappPageRoute), p.Handler(newWebAppEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureSetupWhatsappOTPRoute(webappPageRoute), p.Handler(newWebAppSetupWhatsappOTPHandler))
	router.Add(webapphandler.ConfigureWhatsappOTPRoute(webappPageRoute), p.Handler(newWebAppWhatsappOTPHandler))
	router.Add(webapphandler.ConfigureSetupLoginLinkOTPRoute(webappPageRoute), p.Handler(newWebAppSetupLoginLinkOTPHandler))
	router.Add(webapphandler.ConfigureLoginLinkOTPRoute(webappPageRoute), p.Handler(newWebAppLoginLinkOTPHandler))
	router.Add(webapphandler.ConfigureVerifyLoginLinkOTPRoute(webappPageRoute), p.Handler(newWebAppVerifyLoginLinkOTPHandler))
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
	router.Add(webapphandler.ConfigureFeatureDisabledRoute(webappPageRoute), p.Handler(newWebAppFeatureDisabledHandler))

	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppResetPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountSuccessRoute(webappSuccessPageRoute), p.Handler(newWebAppSettingsDeleteAccountSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappAuthenticatedRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappSettingsRoute), &webapphandler.SettingsImplementationSwitcherHandler{
		SettingV1: p.Handler(newWebAppSettingsHandler),
		SettingV2: p.Handler(newWebAppAuthflowV2SettingsHandler),
	})

	router.Add(webapphandler.ConfigureSettingsProfileRoute(webappSettingsRoute), &webapphandler.SettingsImplementationSwitcherHandler{
		SettingV1: p.Handler(newWebAppSettingsProfileHandler),
		SettingV2: p.Handler(newWebAppAuthflowV2SettingsProfile),
	})
	router.Add(webapphandler.ConfigureSettingsProfileEditRoute(webappSettingsSubRoutesRoute), &webapphandler.SettingsImplementationSwitcherHandler{
		SettingV1: p.Handler(newWebAppSettingsProfileEditHandler),
		SettingV2: p.Handler(newWebAppAuthflowV2SettingsProfileEditHandler),
	})
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityChangePrimaryEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityChangePrimaryEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityAddEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityAddEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityEditEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityEditEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityVerifyEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityVerifyEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityViewEmailRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityViewEmailHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsIdentityListUsername(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsIdentityListUsernameHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsBiometricRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsBiometricHandler))
	router.Add(webapphandler.ConfigureSettingsMFARoute(webappSettingsSubRoutesRoute), &webapphandler.SettingsImplementationSwitcherHandler{
		SettingV1: p.Handler(newWebAppSettingsMFAHandler),
		SettingV2: p.Handler(newWebAppAuthflowV2SettingsMFAHandler),
	})
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFACreatePassword(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFACreatePasswordHandler))
	router.Add(webapphandlerauthflowv2.ConfigureAuthflowV2SettingsMFAPassword(webappSettingsSubRoutesRoute), p.Handler(newWebAppAuthflowV2SettingsMFAPasswordHandler))
	router.Add(webapphandler.ConfigureSettingsTOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsTOTPHandler))
	router.Add(webapphandler.ConfigureSettingsPasskeyRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsPasskeyHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsRecoveryCodeRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSettingsSessionsRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsSessionsHandler))

	router.Add(webapphandler.ConfigureSettingsChangePasswordRoute(webappSettingsSubRoutesRoute), &webapphandler.SettingsImplementationSwitcherHandler{
		SettingV1: p.Handler(newWebAppSettingsChangePasswordHandler),
		SettingV2: p.Handler(newWebAppAuthflowV2SettingsChangePasswordHandler),
	})

	router.Add(webapphandler.ConfigureSettingsChangeSecondaryPasswordRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsChangeSecondaryPasswordHandler))
	router.Add(webapphandler.ConfigureSettingsDeleteAccountRoute(webappSettingsSubRoutesRoute), p.Handler(newWebAppSettingsDeleteAccountHandler))

	router.Add(webapphandler.ConfigureTesterRoute(webappTesterRouter), p.Handler(newWebAppTesterHandler))

	router.Add(webapphandler.ConfigureWechatAuthRoute(webappPageRoute), p.Handler(newWechatAuthHandler))

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

	router.NotFound(webappPageRoute, &webapphandler.ImplementationSwitcherHandler{
		Interaction: p.Handler(newWebAppNotFoundHandler),
		Authflow:    p.Handler(newWebAppNotFoundHandler),
		AuthflowV2:  p.Handler(newWebAppAuthflowV2NotFoundHandler),
	})

	return router
}
