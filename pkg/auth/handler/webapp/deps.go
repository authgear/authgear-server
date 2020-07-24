package webapp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewHTMLRendererLogger,
	wire.Struct(new(HTMLRenderer), "*"),
	wire.Struct(new(FormPrefiller), "*"),
	wire.Bind(new(Renderer), new(*HTMLRenderer)),

	wire.Struct(new(RootHandler), "*"),
	wire.Struct(new(LoginHandler), "*"),
	wire.Struct(new(SignupHandler), "*"),
	wire.Struct(new(PromoteHandler), "*"),
	wire.Struct(new(SSOCallbackHandler), "*"),
	wire.Struct(new(EnterLoginIDHandler), "*"),
	wire.Struct(new(EnterPasswordHandler), "*"),
	wire.Struct(new(CreatePasswordHandler), "*"),
	wire.Struct(new(OOBOTPHandler), "*"),
	wire.Struct(new(ForgotPasswordHandler), "*"),
	wire.Struct(new(ForgotPasswordSuccessHandler), "*"),
	wire.Struct(new(ResetPasswordHandler), "*"),
	wire.Struct(new(ResetPasswordSuccessHandler), "*"),
	wire.Struct(new(SettingsHandler), "*"),
	wire.Struct(new(SettingsIdentityHandler), "*"),
	wire.Struct(new(LogoutHandler), "*"),
)
