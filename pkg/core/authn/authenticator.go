package authn

type AuthenticatorType string

const (
	AuthenticatorTypePassword     AuthenticatorType = "password"
	AuthenticatorTypeTOTP         AuthenticatorType = "totp"
	AuthenticatorTypeOOB          AuthenticatorType = "oob_otp"
	AuthenticatorTypeRecoveryCode AuthenticatorType = "recovery_code"
	AuthenticatorTypeBearerToken  AuthenticatorType = "bearer_token"
)

type AuthenticatorOOBChannel string

const (
	AuthenticatorOOBChannelSMS   AuthenticatorOOBChannel = "sms"
	AuthenticatorOOBChannelEmail AuthenticatorOOBChannel = "email"
)
