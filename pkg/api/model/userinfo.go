package model

import "time"

type UserInfoAuthenticator struct {
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Type      AuthenticatorType `json:"type"`
	Kind      AuthenticatorKind `json:"kind"`

	// oob_otp_sms
	Phone string `json:"phone,omitzero"`

	// oob_otp_email
	Email string `json:"email,omitzero"`

	// totp
	DisplayName string `json:"display_name,omitzero"`
}
