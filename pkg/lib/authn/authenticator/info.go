package authenticator

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Info struct {
	ID        string                  `json:"id"`
	UserID    string                  `json:"user_id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	Type      model.AuthenticatorType `json:"type"`
	Secret    string                  `json:"secret"`
	IsDefault bool                    `json:"is_default"`
	Kind      Kind                    `json:"kind"`
	Claims    map[string]interface{}  `json:"claims"`
}

func (i *Info) ToSpec() Spec {
	return Spec{
		UserID:    i.UserID,
		Type:      i.Type,
		IsDefault: i.IsDefault,
		Kind:      i.Kind,
		Claims:    i.Claims,
	}
}

func (i *Info) ToRef() *Ref {
	return &Ref{
		Meta: model.Meta{
			ID:        i.ID,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
		},
		UserID: i.UserID,
		Type:   i.Type,
	}
}

func (i *Info) GetMeta() model.Meta {
	return model.Meta{
		ID:        i.ID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func (i *Info) AMR() []string {
	switch i.Type {
	case model.AuthenticatorTypePassword:
		return []string{model.AMRPWD}
	case model.AuthenticatorTypeTOTP:
		return []string{model.AMROTP}
	case model.AuthenticatorTypeOOBEmail:
		return []string{model.AMROTP}
	case model.AuthenticatorTypeOOBSMS:
		return []string{model.AMROTP, model.AMRSMS}
	default:
		panic("authenticator: unknown authenticator type: " + i.Type)
	}
}

func (i *Info) Equal(that *Info) bool {
	// Authenticator is equal to each other iff the following holds:

	// They are of the same type.
	sameType := i.Type == that.Type
	if !sameType {
		return false
	}

	switch i.Type {
	case model.AuthenticatorTypePassword:
		// If they are password, they have the same primary/secondary tag.
		return i.Kind == that.Kind
	case model.AuthenticatorTypeTOTP:
		// If they are TOTP, they have the same secret, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		return subtle.ConstantTimeCompare([]byte(i.Secret), []byte(that.Secret)) == 1
	case model.AuthenticatorTypeOOBEmail:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		iEmail := i.Claims[AuthenticatorClaimOOBOTPEmail].(string)
		thatEmail := that.Claims[AuthenticatorClaimOOBOTPEmail].(string)
		return iEmail == thatEmail
	case model.AuthenticatorTypeOOBSMS:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		iPhone := i.Claims[AuthenticatorClaimOOBOTPPhone].(string)
		thatPhone := that.Claims[AuthenticatorClaimOOBOTPPhone].(string)
		return iPhone == thatPhone
	default:
		panic("authenticator: unknown authenticator type: " + i.Type)
	}
}

func (i *Info) StandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	switch i.Type {
	case model.AuthenticatorTypePassword:
		break
	case model.AuthenticatorTypeTOTP:
		break
	case model.AuthenticatorTypeOOBEmail:
		claims[model.ClaimEmail] = i.Claims[AuthenticatorClaimOOBOTPEmail].(string)
	case model.AuthenticatorTypeOOBSMS:
		claims[model.ClaimPhoneNumber] = i.Claims[AuthenticatorClaimOOBOTPPhone].(string)
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}

func (i *Info) CanHaveMFA() bool {
	// No primary authenticator implies no secondary authentication is needed.
	if i == nil {
		return false
	}

	// Only primary authenticator can have MFA.
	if i.Kind != KindPrimary {
		return false
	}

	switch i.Type {
	case model.AuthenticatorTypePassword:
		// password is weak so it can have MFA.
		return true
	case model.AuthenticatorTypeOOBEmail:
		// OTP is weak so it can have MFA.
		return true
	case model.AuthenticatorTypeOOBSMS:
		// OTP is weak so it can have MFA.
		return true
	case model.AuthenticatorTypeTOTP:
		// TOTP was disqualified as primary authenticator very long ago.
		// In case we ever reach here, we treat the situation as no MFA.
		return false
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
}
