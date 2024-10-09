package model

import (
	"errors"
)

type AuthenticatorType string

const (
	AuthenticatorTypePassword        AuthenticatorType = "password"
	AuthenticatorTypePasskey         AuthenticatorType = "passkey"
	AuthenticatorTypeTOTP            AuthenticatorType = "totp"
	AuthenticatorTypeOOBEmail        AuthenticatorType = "oob_otp_email"
	AuthenticatorTypeOOBSMS          AuthenticatorType = "oob_otp_sms"
	AuthenticatorTypeFaceRecognition AuthenticatorType = "face_recognition"
)

type AuthenticatorOOBChannel string

const (
	AuthenticatorOOBChannelSMS      AuthenticatorOOBChannel = "sms"
	AuthenticatorOOBChannelEmail    AuthenticatorOOBChannel = "email"
	AuthenticatorOOBChannelWhatsapp AuthenticatorOOBChannel = "whatsapp"
)

type AuthenticatorKind string

const (
	AuthenticatorKindPrimary   AuthenticatorKind = "primary"
	AuthenticatorKindSecondary AuthenticatorKind = "secondary"
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

func (t AuthenticatorType) ToClaimName() ClaimName {
	switch t {
	case AuthenticatorTypeOOBSMS:
		return ClaimPhoneNumber
	case AuthenticatorTypeOOBEmail:
		return ClaimEmail
	}

	panic("authenticator: incompatible authenticator type: " + t)
}

type Authenticator struct {
	Meta
	UserID    string            `json:"user_id"`
	Type      AuthenticatorType `json:"type"`
	IsDefault bool              `json:"is_default"`
	Kind      AuthenticatorKind `json:"kind"`
}
