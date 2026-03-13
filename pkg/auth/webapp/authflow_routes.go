package webapp

const (
	AuthflowRouteLogin   = "/login"
	AuthflowRouteSignup  = "/signup"
	AuthflowRoutePromote = "/flows/promote_user"
	AuthflowRouteReauth  = "/reauth"
	// AuthflowRouteSignupLogin is login because login page has passkey.
	AuthflowRouteSignupLogin = AuthflowRouteLogin

	AuthflowRouteTerminateOtherSessions = "/authflow/terminate_other_sessions"
	// nolint: gosec
	AuthflowRoutePromptCreatePasskey = "/authflow/prompt_create_passkey"
	AuthflowRouteViewRecoveryCode    = "/authflow/view_recovery_code"
	// nolint: gosec
	AuthflowRouteCreatePassword = "/authflow/create_password"
	// nolint: gosec
	AuthflowRouteChangePassword = "/authflow/change_password"
	// nolint: gosec
	AuthflowRouteEnterPassword     = "/authflow/enter_password"
	AuthflowRouteEnterRecoveryCode = "/authflow/enter_recovery_code"
	AuthflowRouteEnterOOBOTP       = "/authflow/enter_oob_otp"
	AuthflowRouteWhatsappOTP       = "/authflow/whatsapp_otp"
	AuthflowRouteOOBOTPLink        = "/authflow/oob_otp_link"
	AuthflowRouteEnterTOTP         = "/authflow/enter_totp"
	AuthflowRouteSetupTOTP         = "/authflow/setup_totp"
	AuthflowRouteSetupOOBOTP       = "/authflow/setup_oob_otp"
	// nolint: gosec
	AuthflowRouteUsePasskey = "/authflow/use_passkey"
	// nolint: gosec
	AuthflowRouteForgotPassword = "/authflow/forgot_password"
	// nolint: gosec
	AuthflowRouteForgotPasswordOTP = "/authflow/forgot_password/otp"
	// nolint: gosec
	AuthflowRouteForgotPasswordSuccess = "/authflow/forgot_password/success"
	// nolint: gosec
	AuthflowRouteResetPassword = "/authflow/reset_password"
	// nolint: gosec
	AuthflowRouteResetPasswordSuccess = "/authflow/reset_password/success"
	AuthflowRouteWechat               = "/authflow/wechat"

	// The following routes are dead ends.
	AuthflowRouteAccountStatus   = "/authflow/account_status"
	AuthflowRouteNoAuthenticator = "/authflow/no_authenticator"

	AuthflowRouteFinishFlow = "/authflow/finish"
)
