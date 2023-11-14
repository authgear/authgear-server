package authflowclienthandlers

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthflowController), "*"),
	NewAuthflowControllerLogger,

	wire.Struct(new(AuthflowLoginHandler), "*"),
	wire.Struct(new(AuthflowSignupHandler), "*"),
	wire.Struct(new(AuthflowPromoteHandler), "*"),
	wire.Struct(new(AuthflowReauthHandler), "*"),
	wire.Struct(new(AuthflowEnterPasswordHandler), "*"),
	wire.Struct(new(AuthflowEnterOOBOTPHandler), "*"),
	wire.Struct(new(AuthflowCreatePasswordHandler), "*"),
	wire.Struct(new(AuthflowEnterTOTPHandler), "*"),
	wire.Struct(new(AuthflowSetupTOTPHandler), "*"),
	wire.Struct(new(AuthflowViewRecoveryCodeHandler), "*"),
	wire.Struct(new(AuthflowWhatsappOTPHandler), "*"),
	wire.Struct(new(AuthflowOOBOTPLinkHandler), "*"),
	wire.Struct(new(AuthflowChangePasswordHandler), "*"),
	wire.Struct(new(AuthflowUsePasskeyHandler), "*"),
	wire.Struct(new(AuthflowPromptCreatePasskeyHandler), "*"),
	wire.Struct(new(AuthflowEnterRecoveryCodeHandler), "*"),
	wire.Struct(new(AuthflowSetupOOBOTPHandler), "*"),
	wire.Struct(new(AuthflowTerminateOtherSessionsHandler), "*"),
	wire.Struct(new(AuthflowWechatHandler), "*"),
	wire.Struct(new(AuthflowForgotPasswordHandler), "*"),
	wire.Struct(new(AuthflowForgotPasswordSuccessHandler), "*"),
	wire.Struct(new(AuthflowResetPasswordHandler), "*"),
	wire.Struct(new(AuthflowResetPasswordSuccessHandler), "*"),
	wire.Struct(new(AuthflowAccountStatusHandler), "*"),
	wire.Struct(new(AuthflowNoAuthenticatorHandler), "*"),
)
