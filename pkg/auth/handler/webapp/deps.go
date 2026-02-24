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
	wire.Struct(new(EnterPasswordHandler), "*"),
	wire.Struct(new(ConfirmTerminateOtherSessionsHandler), "*"),
	wire.Struct(new(UsePasskeyHandler), "*"),
	wire.Struct(new(CreatePasswordHandler), "*"),
	wire.Struct(new(CreatePasskeyHandler), "*"),
	wire.Struct(new(PromptCreatePasskeyHandler), "*"),
	wire.Struct(new(SetupTOTPHandler), "*"),
	wire.Struct(new(EnterTOTPHandler), "*"),
	wire.Struct(new(SetupOOBOTPHandler), "*"),
	wire.Struct(new(EnterOOBOTPHandler), "*"),
	wire.Struct(new(SetupWhatsappOTPHandler), "*"),
	wire.Struct(new(WhatsappOTPHandler), "*"),
	wire.Struct(new(SetupLoginLinkOTPHandler), "*"),
	wire.Struct(new(LoginLinkOTPHandler), "*"),
	wire.Struct(new(VerifyLoginLinkOTPHandler), "*"),
	wire.Struct(new(EnterRecoveryCodeHandler), "*"),
	wire.Struct(new(SetupRecoveryCodeHandler), "*"),
	wire.Struct(new(VerifyIdentityHandler), "*"),
	wire.Struct(new(VerifyIdentitySuccessHandler), "*"),
	wire.Struct(new(ForgotPasswordHandler), "*"),
	wire.Struct(new(ForgotPasswordSuccessHandler), "*"),
	wire.Struct(new(ResetPasswordHandler), "*"),
	wire.Struct(new(ResetPasswordSuccessHandler), "*"),

	wire.Struct(new(TesterHandler), "*"),

	wire.Struct(new(ForceChangePasswordHandler), "*"),

	wire.Struct(new(ForceChangeSecondaryPasswordHandler), "*"),

	wire.Struct(new(AccountStatusHandler), "*"),
	wire.Struct(new(LogoutHandler), "*"),
	wire.Struct(new(ReturnHandler), "*"),
	wire.Struct(new(WebsocketHandler), "*"),
	wire.Struct(new(WechatAuthHandler), "*"),
	wire.Struct(new(WechatCallbackHandler), "*"),
	wire.Struct(new(PasskeyCreationOptionsHandler), "*"),
	wire.Struct(new(PasskeyRequestOptionsHandler), "*"),
	wire.Struct(new(ConnectWeb3AccountHandler), "*"),
	wire.Struct(new(FeatureDisabledHandler), "*"),

	wire.Struct(new(ResponseWriter), "*"),

	wire.Struct(new(NoProjectSSOCallbackHandler), "*"),
	wire.Struct(new(WhatsappCloudAPIWebhookHandler), "*"),
)
