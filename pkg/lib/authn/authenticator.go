package authn

import "errors"

type AuthenticatorType string

const (
	AuthenticatorTypePassword AuthenticatorType = "password"
	AuthenticatorTypeTOTP     AuthenticatorType = "totp"
	AuthenticatorTypeOOBEmail AuthenticatorType = "oob_otp_email"
	AuthenticatorTypeOOBSMS   AuthenticatorType = "oob_otp_sms"
)

type AuthenticatorOOBChannel string

const (
	AuthenticatorOOBChannelSMS   AuthenticatorOOBChannel = "sms"
	AuthenticatorOOBChannelEmail AuthenticatorOOBChannel = "email"
)

func GetOOBAuthenticatorType(channel AuthenticatorOOBChannel) (AuthenticatorType, error) {
	switch channel {
	case "sms":
		return AuthenticatorTypeOOBSMS, nil
	case "email":
		return AuthenticatorTypeOOBEmail, nil
	default:
		return "", errors.New("invalid oob channel")
	}
}
