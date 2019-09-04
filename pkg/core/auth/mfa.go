package auth

type AuthenticatorType string

const (
	AuthenticatorTypeTOTP         AuthenticatorType = "totp"
	AuthenticatorTypeOOB          AuthenticatorType = "oob"
	AuthenticatorTypeRecoveryCode AuthenticatorType = "recovery_code"
	AuthenticatorTypeBearerToken  AuthenticatorType = "bearer_token"
)

type AuthenticatorOOBChannel string

const (
	AuthenticatorOOBChannelSMS   AuthenticatorOOBChannel = "sms"
	AuthenticatorOOBChannelEmail AuthenticatorOOBChannel = "email"
)
