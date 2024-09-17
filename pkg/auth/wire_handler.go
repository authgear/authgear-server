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
	handlersiwe "github.com/authgear/authgear-server/pkg/auth/handler/siwe"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	handlerwebappauthflowv2 "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2"
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

func newSIWENonceHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlersiwe.NonceHandler)),
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

func newWebAppLoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.LoginHandler)),
	))
}

func newWebAppSignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SignupHandler)),
	))
}

func newWebAppPromoteHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.PromoteHandler)),
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

func newWebAppSSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SSOCallbackHandler)),
	))
}

func newWebAppAuthflowSSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SSOCallbackHandler)),
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

func newWebAppEnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterPasswordHandler)),
	))
}

func newWebConfirmTerminateOtherSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ConfirmTerminateOtherSessionsHandler)),
	))
}

func newWebAppUsePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.UsePasskeyHandler)),
	))
}

func newWebAppCreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.CreatePasswordHandler)),
	))
}

func newWebAppCreatePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.CreatePasskeyHandler)),
	))
}

func newWebAppPromptCreatePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.PromptCreatePasskeyHandler)),
	))
}

func newWebAppSetupTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SetupTOTPHandler)),
	))
}

func newWebAppEnterTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterTOTPHandler)),
	))
}

func newWebAppSetupOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SetupOOBOTPHandler)),
	))
}

func newWebAppEnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterOOBOTPHandler)),
	))
}

func newWebAppSetupWhatsappOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SetupWhatsappOTPHandler)),
	))
}

func newWebAppWhatsappOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WhatsappOTPHandler)),
	))
}

func newWebAppSetupLoginLinkOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SetupLoginLinkOTPHandler)),
	))
}

func newWebAppLoginLinkOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.LoginLinkOTPHandler)),
	))
}

func newWebAppVerifyLoginLinkOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.VerifyLoginLinkOTPHandler)),
	))
}

func newWebAppAuthflowV2VerifyLoginLinkOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2VerifyLoginLinkOTPHandler)),
	))
}

func newWebAppEnterRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.EnterRecoveryCodeHandler)),
	))
}

func newWebAppSetupRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SetupRecoveryCodeHandler)),
	))
}

func newWebAppVerifyIdentityHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.VerifyIdentityHandler)),
	))
}

func newWebAppVerifyIdentitySuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.VerifyIdentitySuccessHandler)),
	))
}

func newWebAppForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForgotPasswordHandler)),
	))
}

func newWebAppForgotPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForgotPasswordSuccessHandler)),
	))
}

func newWebAppResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ResetPasswordHandler)),
	))
}

func newWebAppResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ResetPasswordSuccessHandler)),
	))
}

func newWebAppSettingsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsHandler)),
	))
}

func newWebAppAuthflowV2SettingsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsHandler)),
	))
}

func newWebAppSettingsProfileHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsProfileHandler)),
	))
}

func newWebAppSettingsProfileEditHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsProfileEditHandler)),
	))
}

func newWebAppAuthflowV2SettingsProfileEditHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsProfileEditHandler)),
	))
}

func newWebAppSettingsIdentityHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsIdentityHandler)),
	))
}

func newWebAppSettingsBiometricHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsBiometricHandler)),
	))
}

func newWebAppAuthflowV2SettingsBiometricHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsBiometricHandler)),
	))
}

func newWebAppSettingsMFAHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsMFAHandler)),
	))
}

func newWebAppAuthflowV2SettingsMFAHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsMFAHandler)),
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

func newWebAppSettingsTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsTOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsTOTPHandler)),
	))
}

func newWebAppAuthflowV2SettingsOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsOOBOTPHandler)),
	))
}

func newWebAppSettingsPasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsPasskeyHandler)),
	))
}

func newWebAppAuthflowV2SettingsChangePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsChangePasskeyHandler)),
	))
}

func newWebAppSettingsOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsOOBOTPHandler)),
	))
}

func newWebAppSettingsRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsRecoveryCodeHandler)),
	))
}

func newWebAppSettingsSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsSessionsHandler)),
	))
}

func newWebAppAuthflowV2SettingsSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsSessionsHandler)),
	))
}

func newWebAppForceChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForceChangePasswordHandler)),
	))
}

func newWebAppSettingsChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsChangePasswordHandler)),
	))
}

func newWebAppAuthflowV2SettingsChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SettingsChangePasswordHandler)),
	))
}

func newWebAppForceChangeSecondaryPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ForceChangeSecondaryPasswordHandler)),
	))
}

func newWebAppSettingsChangeSecondaryPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsChangeSecondaryPasswordHandler)),
	))
}

func newWebAppSettingsDeleteAccountHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsDeleteAccountHandler)),
	))
}

func newWebAppSettingsDeleteAccountSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsDeleteAccountSuccessHandler)),
	))
}

func newWebAppAccountStatusHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AccountStatusHandler)),
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

func newWebAppErrorHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ErrorHandler)),
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

func newWebAppNotFoundHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.NotFoundHandler)),
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

func newWebAppConnectWeb3AccountHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ConnectWeb3AccountHandler)),
	))
}

func newWebAppMissingWeb3WalletHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.MissingWeb3WalletHandler)),
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

func newWebAppAuthflowLoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowLoginHandler)),
	))
}

func newWebAppAuthflowV2LoginHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2LoginHandler)),
	))
}

func newWebAppAuthflowSignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowSignupHandler)),
	))
}

func newWebAppAuthflowV2SignupHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SignupHandler)),
	))
}

func newWebAppAuthflowPromoteHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowPromoteHandler)),
	))
}

func newWebAppAuthflowV2PromoteHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2PromoteHandler)),
	))
}

func newWebAppAuthflowEnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowEnterPasswordHandler)),
	))
}

func newWebAppAuthflowV2EnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterPasswordHandler)),
	))
}

func newWebAppAuthflowEnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowEnterOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2EnterOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterOOBOTPHandler)),
	))
}

func newWebAppAuthflowCreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowCreatePasswordHandler)),
	))
}

func newWebAppAuthflowV2CreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2CreatePasswordHandler)),
	))
}

func newWebAppAuthflowEnterTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowEnterTOTPHandler)),
	))
}

func newWebAppAuthflowV2EnterTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterTOTPHandler)),
	))
}

func newWebAppAuthflowSetupTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowSetupTOTPHandler)),
	))
}

func newWebAppAuthflowV2SetupTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SetupTOTPHandler)),
	))
}

func newWebAppAuthflowViewRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowViewRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowV2ViewRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ViewRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowWhatsappOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowWhatsappOTPHandler)),
	))
}

func newWebAppAuthflowOOBOTPLinkHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowOOBOTPLinkHandler)),
	))
}

func newWebAppAuthflowV2OOBOTPLinkHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2OOBOTPLinkHandler)),
	))
}

func newWebAppAuthflowChangePasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowChangePasswordHandler)),
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

func newWebAppAuthflowUsePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowUsePasskeyHandler)),
	))
}

func newWebAppAuthflowV2UsePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2UsePasskeyHandler)),
	))
}

func newWebAppAuthflowPromptCreatePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowPromptCreatePasskeyHandler)),
	))
}

func newWebAppAuthflowV2PromptCreatePasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2PromptCreatePasskeyHandler)),
	))
}

func newWebAppAuthflowEnterRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowEnterRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowV2EnterRecoveryCodeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2EnterRecoveryCodeHandler)),
	))
}

func newWebAppAuthflowSetupOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowSetupOOBOTPHandler)),
	))
}

func newWebAppAuthflowV2SetupOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2SetupOOBOTPHandler)),
	))
}

func newWebAppAuthflowTerminateOtherSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowTerminateOtherSessionsHandler)),
	))
}

func newWebAppAuthflowV2TerminateOtherSessionsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2TerminateOtherSessionsHandler)),
	))
}

func newWebAppAuthflowWechatHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowWechatHandler)),
	))
}

func newWebAppAuthflowForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowForgotPasswordHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordHandler)),
	))
}

func newWebAppAuthflowForgotPasswordOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowForgotPasswordOTPHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordOTPHandler)),
	))
}

func newWebAppAuthflowForgotPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowForgotPasswordSuccessHandler)),
	))
}

func newWebAppAuthflowV2ForgotPasswordLinkSentHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ForgotPasswordLinkSentHandler)),
	))
}

func newWebAppReauthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.ReauthHandler)),
	))
}

func newWebAppAuthflowReauthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowReauthHandler)),
	))
}

func newWebAppAuthflowV2ReauthHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ReauthHandler)),
	))
}

func newWebAppAuthflowResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowResetPasswordHandler)),
	))
}

func newWebAppAuthflowV2ResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ResetPasswordHandler)),
	))
}

func newWebAppAuthflowResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowResetPasswordSuccessHandler)),
	))
}

func newWebAppAuthflowV2ResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowV2UIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebappauthflowv2.AuthflowV2ResetPasswordSuccessHandler)),
	))
}

func newWebAppAuthflowAccountStatusHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowAccountStatusHandler)),
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
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowNoAuthenticatorHandler)),
	))
}

func newWebAppAuthflowFinishFlowHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		AuthflowUIHandlerDependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.AuthflowFinishFlowHandler)),
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
