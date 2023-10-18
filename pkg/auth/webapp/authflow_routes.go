package webapp

const (
	AuthflowRouteLogin  = "/login"
	AuthflowRouteSignup = "/signup"
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
	AuthflowRouteUsePasskey    = "/authflow/use_passkey"
	AuthflowRouteAccountStatus = "/authflow/account_status"
)
