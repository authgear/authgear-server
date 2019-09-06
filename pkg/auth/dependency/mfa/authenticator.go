package mfa

import (
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
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

func CanAddAuthenticator(authenticators []interface{}, newA interface{}, mfaConfiguration config.MFAConfiguration) bool {
	// Calculate the count
	totalCount := len(authenticators)
	totpCount := 0
	oobSMSCount := 0
	oobEmailCount := 0
	for _, a := range authenticators {
		switch aa := a.(type) {
		case TOTPAuthenticator:
			totpCount++
		case OOBAuthenticator:
			switch aa.Channel {
			case coreAuth.AuthenticatorOOBChannelSMS:
				oobSMSCount++
			case coreAuth.AuthenticatorOOBChannelEmail:
				oobSMSCount++
			default:
				panic("unknown OOB authenticator channel")
			}
		default:
			panic("unknown authenticator")
		}
	}

	// Simulate the count if new one is added.
	totalCount++
	switch newAA := newA.(type) {
	case TOTPAuthenticator:
		totpCount++
	case OOBAuthenticator:
		switch newAA.Channel {
		case coreAuth.AuthenticatorOOBChannelSMS:
			oobSMSCount++
		case coreAuth.AuthenticatorOOBChannelEmail:
			oobSMSCount++
		default:
			panic("unknown OOB authenticator channel")
		}
	}

	// Compare the count
	if totalCount > *mfaConfiguration.Maximum {
		return false
	}
	if totpCount > mfaConfiguration.TOTP.Maximum {
		return false
	}
	if oobSMSCount > mfaConfiguration.OOB.SMS.Maximum {
		return false
	}
	if oobEmailCount > mfaConfiguration.OOB.Email.Maximum {
		return false
	}

	return true
}

func IsDeletingLastActivatedAuthenticator(authenticators []interface{}, a interface{}) bool {
	id := ""
	activated := false
	switch a := a.(type) {
	case TOTPAuthenticator:
		id = a.ID
		activated = a.Activated
	case OOBAuthenticator:
		id = a.ID
		activated = a.Activated
	default:
		panic("unknown authenticator")
	}

	if !activated {
		return false
	}

	if len(authenticators) != 1 {
		return false
	}

	for _, aa := range authenticators {
		switch aa := aa.(type) {
		case TOTPAuthenticator:
			if aa.ID == id {
				return true
			}
		case OOBAuthenticator:
			if aa.ID == id {
				return true
			}
		default:
			panic("unknown authenticator")
		}
	}
	return false
}
