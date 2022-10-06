//go:build wireinject
// +build wireinject

package auth

import (
	"context"
	"net/http"

	"github.com/google/wire"

	handlerapi "github.com/authgear/authgear-server/pkg/auth/handler/api"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlersiwe "github.com/authgear/authgear-server/pkg/auth/handler/siwe"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
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

func newWebAppSSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
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

func newWhatsappWATICallbackHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.WhatsappWATICallbackHandler)),
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

func newWebAppSettingsMFAHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsMFAHandler)),
	))
}

func newWebAppSettingsTOTPHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsTOTPHandler)),
	))
}

func newWebAppSettingsPasskeyHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handlerwebapp.SettingsPasskeyHandler)),
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
