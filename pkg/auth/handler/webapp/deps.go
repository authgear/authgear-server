package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthEntryPointMiddleware), "*"),
	wire.Struct(new(ImplementationSwitcherMiddleware), "*"),
	wire.Struct(new(SettingsImplementationSwitcherMiddleware), "*"),

	wire.Struct(new(ResponseRenderer), "*"),
	wire.Struct(new(FormPrefiller), "*"),
	wire.Bind(new(Renderer), new(*ResponseRenderer)),
	wire.Struct(new(ErrorRenderer), "*"),

	wire.Struct(new(ControllerDeps), "*"),
	wire.Struct(new(ControllerFactory), "*"),

	wire.Struct(new(AuthflowController), "*"),

	NewPublisher,
	wire.Struct(new(GlobalSessionServiceFactory), "*"),

	wire.Struct(new(PanicMiddleware), "*"),

	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(CSRFErrorInstructionHandler), "*"),

	wire.Struct(new(AppStaticAssetsHandler), "*"),

	wire.Struct(new(RootHandler), "*"),
	wire.Struct(new(OAuthEntrypointHandler), "*"),
	wire.Struct(new(SelectAccountHandler), "*"),
	wire.Struct(new(SSOCallbackHandler), "*"),
	wire.Struct(new(EnterLoginIDHandler), "*"),
	wire.Struct(new(ForgotPasswordSuccessHandler), "*"),
	wire.Struct(new(ResetPasswordSuccessHandler), "*"),

	wire.Struct(new(TesterHandler), "*"),

	wire.Struct(new(LogoutHandler), "*"),
	wire.Struct(new(ReturnHandler), "*"),
	wire.Struct(new(WebsocketHandler), "*"),
	wire.Struct(new(WechatAuthHandler), "*"),
	wire.Struct(new(WechatCallbackHandler), "*"),
	wire.Struct(new(PasskeyCreationOptionsHandler), "*"),
	wire.Struct(new(PasskeyRequestOptionsHandler), "*"),
	wire.Struct(new(FeatureDisabledHandler), "*"),

	wire.Struct(new(ResponseWriter), "*"),

	wire.Struct(new(NoProjectSSOCallbackHandler), "*"),
	wire.Struct(new(WhatsappCloudAPIWebhookHandler), "*"),
)
