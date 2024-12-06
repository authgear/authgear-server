package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthEntryPointMiddleware), "*"),

	wire.Struct(new(ResponseRenderer), "*"),
	wire.Struct(new(FormPrefiller), "*"),
	wire.Bind(new(Renderer), new(*ResponseRenderer)),
	wire.Struct(new(ErrorRenderer), "*"),

	wire.Struct(new(ControllerDeps), "*"),
	wire.Struct(new(ControllerFactory), "*"),

	wire.Struct(new(AuthflowController), "*"),
	NewAuthflowControllerLogger,

	NewPublisher,
	wire.Struct(new(GlobalSessionServiceFactory), "*"),

	NewPanicMiddlewareLogger,
	wire.Struct(new(PanicMiddleware), "*"),

	NewCSRFMiddlewareLogger,
	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(CSRFErrorInstructionHandler), "*"),

	wire.Struct(new(AppStaticAssetsHandler), "*"),

	wire.Struct(new(RootHandler), "*"),
	wire.Struct(new(OAuthEntrypointHandler), "*"),
	wire.Struct(new(SSOCallbackHandler), "*"),
	wire.Struct(new(TesterHandler), "*"),
	wire.Struct(new(LogoutHandler), "*"),
	wire.Struct(new(ReturnHandler), "*"),
	wire.Struct(new(ErrorHandler), "*"),
	wire.Struct(new(WebsocketHandler), "*"),
	wire.Struct(new(WechatCallbackHandler), "*"),
	wire.Struct(new(PasskeyCreationOptionsHandler), "*"),
	wire.Struct(new(PasskeyRequestOptionsHandler), "*"),
	wire.Struct(new(FeatureDisabledHandler), "*"),

	wire.Struct(new(ResponseWriter), "*"),
)
