package webapp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewResponseRendererLogger,
	wire.Struct(new(ResponseRenderer), "*"),
	wire.Struct(new(FormPrefiller), "*"),
	wire.Bind(new(Renderer), new(*ResponseRenderer)),

	wire.Struct(new(PanicMiddleware), "*"),

	wire.Struct(new(StaticAssetsHandler), "*"),

	wire.Struct(new(RootHandler), "*"),
	wire.Struct(new(LoginHandler), "*"),
	wire.Struct(new(SignupHandler), "*"),
	wire.Struct(new(PromoteHandler), "*"),
	wire.Struct(new(SSOCallbackHandler), "*"),
	wire.Struct(new(EnterLoginIDHandler), "*"),
	wire.Struct(new(EnterPasswordHandler), "*"),
	wire.Struct(new(CreatePasswordHandler), "*"),
	wire.Struct(new(SetupTOTPHandler), "*"),
	wire.Struct(new(EnterTOTPHandler), "*"),
	wire.Struct(new(SetupOOBOTPHandler), "*"),
	wire.Struct(new(EnterOOBOTPHandler), "*"),
	wire.Struct(new(EnterRecoveryCodeHandler), "*"),
	wire.Struct(new(SetupRecoveryCodeHandler), "*"),
	wire.Struct(new(VerifyIdentityHandler), "*"),
	wire.Struct(new(VerifyIdentitySuccessHandler), "*"),
	wire.Struct(new(ForgotPasswordHandler), "*"),
	wire.Struct(new(ForgotPasswordSuccessHandler), "*"),
	wire.Struct(new(ResetPasswordHandler), "*"),
	wire.Struct(new(ResetPasswordSuccessHandler), "*"),
	wire.Struct(new(SettingsHandler), "*"),
	wire.Struct(new(SettingsIdentityHandler), "*"),
	wire.Struct(new(SettingsTOTPHandler), "*"),
	wire.Struct(new(SettingsOOBOTPHandler), "*"),
	wire.Struct(new(SettingsRecoveryCodeHandler), "*"),
	wire.Struct(new(ChangePasswordHandler), "*"),
	wire.Struct(new(ChangeSecondaryPasswordHandler), "*"),
	wire.Struct(new(LogoutHandler), "*"),
	wire.Struct(new(AuthenticationBeginHandler), "*"),
	wire.Struct(new(CreateAuthenticatorBeginHandler), "*"),
)
