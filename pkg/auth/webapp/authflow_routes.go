package webapp

const (
	AuthflowRouteLogin  = "/authflow/login"
	AuthflowRouteSignup = "/authflow/signup"
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
	AuthflowRouteEnterTOTP         = "/authflow/enter_totp"
	AuthflowRouteSetupTOTP         = "/authflow/setup_totp"
	AuthflowRouteWhatsappOTP       = "/authflow/whatsapp_otp"
	AuthflowRouteOOBOTPLink        = "/authflow/oob_otp_link"
	// nolint: gosec
	AuthflowRouteUsePasskey = "/authflow/use_passkey"
)
