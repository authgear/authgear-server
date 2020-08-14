package authn

type AuthenticatorType string

const (
	AuthenticatorTypePassword AuthenticatorType = "password"
	AuthenticatorTypeTOTP     AuthenticatorType = "totp"
	AuthenticatorTypeOOB      AuthenticatorType = "oob_otp"
)

type AuthenticatorOOBChannel string

const (
	AuthenticatorOOBChannelSMS   AuthenticatorOOBChannel = "sms"
	AuthenticatorOOBChannelEmail AuthenticatorOOBChannel = "email"
)
