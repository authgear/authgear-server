package authenticator

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type Info struct {
	ID        string                  `json:"id"`
	UserID    string                  `json:"user_id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	Type      model.AuthenticatorType `json:"type"`
	IsDefault bool                    `json:"is_default"`
	Kind      Kind                    `json:"kind"`

	Password *Password `json:"password,omitempty"`
	Passkey  *Passkey  `json:"passkey,omitempty"`
	TOTP     *TOTP     `json:"totp,omitempty"`
	OOBOTP   *OOBOTP   `json:"oobotp,omitempty"`
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
	case model.AuthenticatorTypePasskey:
		return []string{model.AMRXPasskey}
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
	case model.AuthenticatorTypePasskey:
		// if they are passkey, they have the same credential ID.
		return i.Passkey.CredentialID == that.Passkey.CredentialID
	case model.AuthenticatorTypeTOTP:
		// If they are TOTP, they have the same secret, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}
		iSecret := i.TOTP.Secret
		thatSecret := that.TOTP.Secret
		return subtle.ConstantTimeCompare([]byte(iSecret), []byte(thatSecret)) == 1
	case model.AuthenticatorTypeOOBEmail:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}
		return i.OOBOTP.Email == that.OOBOTP.Email
	case model.AuthenticatorTypeOOBSMS:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		return i.OOBOTP.Phone == that.OOBOTP.Phone
	default:
		panic("authenticator: unknown authenticator type: " + i.Type)
	}
}

func (i *Info) ToPublicClaims() map[string]interface{} {
	claims := make(map[string]interface{})
	switch i.Type {
	case model.AuthenticatorTypeTOTP:
		claims[AuthenticatorClaimTOTPDisplayName] = i.TOTP.DisplayName
	case model.AuthenticatorTypeOOBEmail:
		claims[AuthenticatorClaimOOBOTPEmail] = i.OOBOTP.Email
	case model.AuthenticatorTypeOOBSMS:
		claims[AuthenticatorClaimOOBOTPPhone] = i.OOBOTP.Phone
	case model.AuthenticatorTypePasskey:
		claims[AuthenticatorClaimPasskeyCredentialID] = i.Passkey.CredentialID
	default:
		// no claims to add
		break
	}
	return claims
}

func (i *Info) StandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	switch i.Type {
	case model.AuthenticatorTypePassword:
		break
	case model.AuthenticatorTypePasskey:
		break
	case model.AuthenticatorTypeTOTP:
		break
	case model.AuthenticatorTypeOOBEmail:
		claims[model.ClaimEmail] = i.OOBOTP.Email
	case model.AuthenticatorTypeOOBSMS:
		claims[model.ClaimPhoneNumber] = i.OOBOTP.Phone
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
	case model.AuthenticatorTypePasskey:
		// passkey is strong so it cannot have MFA.
		return false
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

func (i *Info) IsIndependent() bool {
	switch i.Kind {
	case KindPrimary:
		switch i.Type {
		case model.AuthenticatorTypeOOBEmail,
			model.AuthenticatorTypeOOBSMS,
			model.AuthenticatorTypePasskey:
			return false
		default:
			return true
		}
	case KindSecondary:
		return true
	default:
		panic(fmt.Errorf("authenticator: unknown kind: %v", i.Kind))
	}
}

func (i *Info) IsDependentOf(iden *identity.Info) bool {
	// Primary OOB OTP authenticator is the dependent of Login ID identity.
	if i.Kind == KindPrimary && (i.Type == model.AuthenticatorTypeOOBEmail || i.Type == model.AuthenticatorTypeOOBSMS) {
		identityClaims := iden.IdentityAwareStandardClaims()
		for k, v := range i.StandardClaims() {
			if iden.Type == model.IdentityTypeLoginID && identityClaims[k] == v {
				return true
			}
		}
	}

	// primary passkey is the dependent of passkey identity.
	if i.Kind == KindPrimary && i.Type == model.AuthenticatorTypePasskey {
		if iden.Type == model.IdentityTypePasskey {
			if i.Passkey.CredentialID == iden.Passkey.CredentialID {
				return true
			}
		}
	}

	return false
}

func (i *Info) IsApplicableTo(iden *identity.Info) bool {
	return KeepPrimaryAuthenticatorOfIdentity(iden).Keep(i)
}

func (i *Info) ToModel() model.Authenticator {
	return model.Authenticator{
		Meta:      i.GetMeta(),
		UserID:    i.UserID,
		Type:      i.Type,
		IsDefault: i.IsDefault,
		Kind:      model.AuthenticatorKind(i.Kind),
	}
}

func (i *Info) UpdateUserID(newUserID string) *Info {
	i.UserID = newUserID
	switch i.Type {
	case model.AuthenticatorTypePassword:
		i.Password.UserID = newUserID
	case model.AuthenticatorTypePasskey:
		i.Passkey.UserID = newUserID
	case model.AuthenticatorTypeTOTP:
		i.TOTP.UserID = newUserID
	case model.AuthenticatorTypeOOBEmail:
		fallthrough
	case model.AuthenticatorTypeOOBSMS:
		i.OOBOTP.UserID = newUserID
	default:
		panic(fmt.Errorf("identity: identity type %v does not support updating user ID", i.Type))
	}
	return i
}
