package mfa

import (
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/phone"
)

type TOTPAuthenticator struct {
	ID          string
	UserID      string
	Type        coreAuth.AuthenticatorType
	Activated   bool
	CreatedAt   time.Time
	ActivatedAt *time.Time
	Secret      string
	DisplayName string
}

func (a *TOTPAuthenticator) Mask() MaskedTOTPAuthenticator {
	return MaskedTOTPAuthenticator{
		ID:          a.ID,
		Type:        a.Type,
		CreatedAt:   a.CreatedAt,
		ActivatedAt: a.ActivatedAt,
		DisplayName: a.DisplayName,
	}
}

type MaskedTOTPAuthenticator struct {
	ID          string                     `json:"id"`
	Type        coreAuth.AuthenticatorType `json:"type"`
	CreatedAt   time.Time                  `json:"created_at"`
	ActivatedAt *time.Time                 `json:"activated_at"`
	DisplayName string                     `json:"display_name"`
}

type OOBAuthenticator struct {
	ID          string
	UserID      string
	Type        coreAuth.AuthenticatorType
	Activated   bool
	CreatedAt   time.Time
	ActivatedAt *time.Time
	Channel     coreAuth.AuthenticatorOOBChannel
	Phone       string
	Email       string
}

func (a *OOBAuthenticator) Mask() MaskedOOBAuthenticator {
	return MaskedOOBAuthenticator{
		ID:          a.ID,
		Type:        a.Type,
		CreatedAt:   a.CreatedAt,
		ActivatedAt: a.ActivatedAt,
		Channel:     a.Channel,
		MaskedPhone: phone.Mask(a.Phone),
		MaskedEmail: mail.MaskAddress(a.Email),
	}
}

type MaskedOOBAuthenticator struct {
	ID          string                           `json:"id"`
	Type        coreAuth.AuthenticatorType       `json:"type"`
	CreatedAt   time.Time                        `json:"created_at"`
	ActivatedAt *time.Time                       `json:"activated_at"`
	Channel     coreAuth.AuthenticatorOOBChannel `json:"channel"`
	MaskedPhone string                           `json:"masked_phone,omitempty"`
	MaskedEmail string                           `json:"masked_email,omitempty"`
}

type RecoveryCodeAuthenticator struct {
	ID        string
	UserID    string
	Type      coreAuth.AuthenticatorType
	Code      string
	CreatedAt time.Time
	Consumed  bool
}

type BearerTokenAuthenticator struct {
	ID        string
	UserID    string
	Type      coreAuth.AuthenticatorType
	ParentID  string
	Token     string
	CreatedAt time.Time
	ExpireAt  time.Time
}

type OOBCode struct {
	ID              string
	UserID          string
	AuthenticatorID string
	Code            string
	CreatedAt       time.Time
	ExpireAt        time.Time
}

func MaskAuthenticators(authenticators []interface{}) []interface{} {
	output := make([]interface{}, len(authenticators))
	for i, a := range authenticators {
		switch aa := a.(type) {
		case TOTPAuthenticator:
			output[i] = aa.Mask()
		case OOBAuthenticator:
			output[i] = aa.Mask()
		default:
			panic("unknown authenticator")
		}
	}
	return output
}
