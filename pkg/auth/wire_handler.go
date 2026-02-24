//go:build wireinject
// +build wireinject

package auth

import (
	"context"
	"net/http"

	"github.com/google/wire"

	handlerapi "github.com/authgear/authgear-server/pkg/auth/handler/api"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlersaml "github.com/authgear/authgear-server/pkg/auth/handler/saml"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	handlerwebappauthflowv2 "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/web"
)

func newHealthzHandler(p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	panic(wire.Build(
		deps.RootDependencySet,
		healthz.DependencySet,
		wire.Bind(new(http.Handler), new(*healthz.Handler)),
	))
}

func newWebAppGeneratedStaticAssetsHandler(p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.Struct(new(handlerwebapp.GeneratedStaticAssetsHandler), "*"),
		wire.Bind(new(handlerwebapp.GlobalEmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
		wire.Bind(new(http.Handler), new(*handlerwebapp.GeneratedStaticAssetsHandler)),
	))
}

func newPreviewWidgetHandler(p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	panic(wire.Build(
		NoProjectDependencySet,
		wire.Struct(new(handlerwebappauthflowv2.PreviewWidgetHandler), "*"),
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.PreviewWidgetHandler)),
	))
}

func newNoProjectSSOCallbackHandler(configSource *configsource.ConfigSource, p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	panic(wire.Build(
		NoProjectDependencySet,
		wire.Struct(new(handlerwebapp.NoProjectSSOCallbackHandler), "*"),
		wire.Bind(new(http.Handler), new(*handlerwebapp.NoProjectSSOCallbackHandler)),
	))
}

func newWhatsappCloudAPIWebhookHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WhatsappCloudAPIWebhookHandler)),
	))
}

func newOAuthAuthorizeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.AuthorizeHandler)),
	))
}

func newOAuthConsentHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.ConsentHandler)),
	))
}

func newOAuthTokenHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.TokenHandler)),
	))
}

func newOAuthRevokeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.RevokeHandler)),
	))
}

func newOAuthMetadataHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.MetadataHandler)),
	))
}

func newOAuthJWKSHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.JWKSHandler)),
	))
}

func newOAuthUserInfoHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.UserInfoHandler)),
	))
}

func newOAuthEndSessionHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.EndSessionHandler)),
	))
}

func newOAuthChallengeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.ChallengeHandler)),
	))
}

func newOAuthAppSessionTokenHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.AppSessionTokenHandler)),
	))
}

func newOAuthProxyRedirectHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.ProxyRedirectHandler)),
	))
}

func newAPIAnonymousUserSignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AnonymousUserSignupAPIHandler)),
	))
}

func newAPIAnonymousUserPromotionCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AnonymousUserPromotionCodeAPIHandler)),
	))
}

func newAPIPresignImagesUploadHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.PresignImagesUploadHandler)),
	))
}

func newWebAppOAuthEntrypointHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.OAuthEntrypointHandler)),
	))
}

func newWebAppRootHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.RootHandler)),
	))
}

func newWebAppSelectAccountHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SelectAccountHandler)),
	))
}

func newWebAppAuthflowV2VerifyBotProtectionHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2VerifyBotProtectionHandler)),
	))
}

func newWebAppAuthflowV2SelectAccountHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SelectAccountHandler)),
	))
}

func newWebAppAuthflowV2SSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SSOCallbackHandler)),
	))
}

func newWechatAuthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WechatAuthHandler)),
	))
}

func newWechatCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WechatCallbackHandler)),
	))
}

func newWebAppEnterLoginIDHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterLoginIDHandler)),
	))
}

func newWebAppAuthflowV2VerifyLoginLinkOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2VerifyLoginLinkOTPHandler)),
	))
}

func newWebAppForgotPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForgotPasswordSuccessHandler)),
	))
}

func newWebAppResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ResetPasswordSuccessHandler)),
	))
}

func newWebAppAuthflowV2SettingsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsHandler)),
	))
}

func newWebAppAuthflowV2SettingsProfileEditHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsProfileEditHandler)),
	))
}

func newWebAppAuthflowV2SettingsBiometricHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsBiometricHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAViewRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAViewRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFACreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFACreatePasswordHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAPasswordHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAChangePasswordHandler)),
	))
}

func newWebAppAuthflowV2SettingsTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsTOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFACreateTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFACreateTOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAEnterTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAEnterTOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFACreateOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFACreateOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAEnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAEnterOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsChangePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsChangePasskeyHandler)),
	))
}

func newWebAppAuthflowV2SettingsSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsSessionsHandler)),
	))
}

func newWebAppAuthflowV2SettingsChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsChangePasswordHandler)),
	))
}

func newWebAppAuthflowV2SettingsDeleteAccountHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsDeleteAccountHandler)),
	))
}

func newWebAppAuthflowV2SettingsDeleteAccountSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsDeleteAccountSuccessHandler)),
	))
}

func newWebAppAuthflowV2SettingsAdvancedSettingsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsAdvancedSettingsHandler)),
	))
}

func newWebAppLogoutHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.LogoutHandler)),
	))
}

func newWebAppAppStaticAssetsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AppStaticAssetsHandler)),
	))
}

func newWebAppReturnHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ReturnHandler)),
	))
}

func newWebAppAuthflowV2ErrorHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ErrorHandler)),
	))
}

func newWebAppCSRFErrorInstructionHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.CSRFErrorInstructionHandler)),
	))
}

func newWebAppAuthflowV2NotFoundHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2NotFoundHandler)),
	))
}

func newWebAppWebsocketHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WebsocketHandler)),
	))
}

func newWebAppPasskeyCreationOptionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.PasskeyCreationOptionsHandler)),
	))
}

func newWebAppPasskeyRequestOptionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.PasskeyRequestOptionsHandler)),
	))
}

func newWebAppFeatureDisabledHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.FeatureDisabledHandler)),
	))
}

func newWebAppTesterHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.TesterHandler)),
	))
}

func newAPIWorkflowNewHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.WorkflowNewHandler)),
	))
}

func newAPIWorkflowGetHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.WorkflowGetHandler)),
	))
}

func newAPIWorkflowInputHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.WorkflowInputHandler)),
	))
}

func newAPIWorkflowWebsocketHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.WorkflowWebsocketHandler)),
	))
}

func newAPIWorkflowV2Handler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.WorkflowV2Handler)),
	))
}

func newAPIAuthenticationFlowV1CreateHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AuthenticationFlowV1CreateHandler)),
	))
}

func newAPIAuthenticationFlowV1InputHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AuthenticationFlowV1InputHandler)),
	))
}

func newAPIAuthenticationFlowV1GetHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AuthenticationFlowV1GetHandler)),
	))
}

func newAPIAuthenticationFlowV1WebsocketHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AuthenticationFlowV1WebsocketHandler)),
	))
}

func newAPIAccountManagementV1IdentificationHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AccountManagementV1IdentificationHandler)),
	))
}

func newAPIAccountManagementV1IdentificationOAuthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerapi.AccountManagementV1IdentificationOAuthHandler)),
	))
}

func newWebAppAuthflowV2LoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2LoginHandler)),
	))
}

func newWebAppAuthflowV2SignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SignupHandler)),
	))
}

func newWebAppAuthflowV2PromoteHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2PromoteHandler)),
	))
}

func newWebAppAuthflowV2EnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterPasswordHandler)),
	))
}

func newWebAppAuthflowV2EnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2CreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2CreatePasswordHandler)),
	))
}

func newWebAppAuthflowV2EnterTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterTOTPHandler)),
	))
}

func newWebAppAuthflowV2SetupTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SetupTOTPHandler)),
	))
}

func newWebAppAuthflowV2ViewRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ViewRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowV2OOBOTPLinkHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2OOBOTPLinkHandler)),
	))
}

func newWebAppAuthflowV2ChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ChangePasswordHandler)),
	))
}

func newWebAppAuthflowV2ChangePasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ChangePasswordSuccessHandler)),
	))
}

func newWebAppAuthflowV2UsePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2UsePasskeyHandler)),
	))
}

func newWebAppAuthflowV2PromptCreatePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2PromptCreatePasskeyHandler)),
	))
}

func newWebAppAuthflowV2EnterRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowV2SetupOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SetupOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2TerminateOtherSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2TerminateOtherSessionsHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordOTPHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordLinkSentHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordLinkSentHandler)),
	))
}

func newWebAppAuthflowV2ReauthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ReauthHandler)),
	))
}

func newWebAppAuthflowV2ResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ResetPasswordHandler)),
	))
}

func newWebAppAuthflowV2ResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ResetPasswordSuccessHandler)),
	))
}

func newWebAppAuthflowV2AccountStatusHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2AccountStatusHandler)),
	))
}

func newWebAppAuthflowNoAuthenticatorHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2NoAuthenticatorHandler)),
	))
}

func newWebAppAuthflowV2OAuthProviderMissingCredentialsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2OAuthProviderMissingCredentialsHandler)),
	))
}

func newWebAppAuthflowV2OAuthProviderDemoCredentialHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2OAuthProviderDemoCredentialHandler)),
	))
}

func newWebAppAuthflowV2FinishFlowHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2FinishFlowHandler)),
	))
}

func newWebAppAuthflowV2AccountLinkingHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2AccountLinkingHandler)),
	))
}

func newWebAppAuthflowV2NoAuthenticatorHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2NoAuthenticatorHandler)),
	))
}

func newWebAppAuthflowV2WechatHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2WechatHandler)),
	))
}

func newWebAppAuthflowV2LDAPLoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2LDAPLoginHandler)),
	))
}

func newSAMLMetadataHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlersaml.MetadataHandler)),
	))
}

func newSAMLLoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlersaml.LoginHandler)),
	))
}

func newSAMLLoginFinishHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlersaml.LoginFinishHandler)),
	))
}

func newSAMLLogoutHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlersaml.LogoutHandler)),
	))
}

func newWebAppAuthflowV2SettingsProfile(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsProfileHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityAddEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityAddEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityEditEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityEditEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityListEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityListEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityVerifyEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityVerifyEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityViewEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityViewEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityChangePrimaryEmailHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityChangePrimaryEmailHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityAddPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityAddPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityEditPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityEditPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityListPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityListPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityViewPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityViewPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityChangePrimaryPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityChangePrimaryPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityVerifyPhoneHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityVerifyPhoneHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityListUsernameHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityListUsernameHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityNewUsernameHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityNewUsernameHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityViewUsernameHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityViewUsernameHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityEditUsernameHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityEditUsernameHandler)),
	))
}

func newWebAppAuthflowV2SettingsIdentityListOAuthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsIdentityListOAuthHandler)),
	))
}
