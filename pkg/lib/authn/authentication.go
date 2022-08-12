package authn

type AuthenticationType string

const (
	AuthenticationTypeNone         AuthenticationType = "none"
	AuthenticationTypePassword     AuthenticationType = "password"
	AuthenticationTypePasskey      AuthenticationType = "passkey"
	AuthenticationTypeTOTP         AuthenticationType = "totp"
	AuthenticationTypeOOBOTPEmail  AuthenticationType = "oob_otp_email"
	AuthenticationTypeOOBOTPSMS    AuthenticationType = "oob_otp_sms"
	AuthenticationTypeRecoveryCode AuthenticationType = "recovery_code"
	AuthenticationTypeDeviceToken  AuthenticationType = "device_token"
)

type AuthenticationStage string

const (
	AuthenticationStagePrimary   AuthenticationStage = "primary"
	AuthenticationStageSecondary AuthenticationStage = "secondary"
)
