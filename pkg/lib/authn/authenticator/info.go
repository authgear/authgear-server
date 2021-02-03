package authenticator

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Info struct {
	ID        string                  `json:"id"`
	Labels    map[string]interface{}  `json:"labels"`
	UserID    string                  `json:"user_id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	Type      authn.AuthenticatorType `json:"type"`
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
	case authn.AuthenticatorTypePassword:
		return []string{authn.AMRPWD}
	case authn.AuthenticatorTypeTOTP:
		return []string{authn.AMROTP}
	case authn.AuthenticatorTypeOOBEmail, authn.AuthenticatorTypeOOBSMS:
		out := []string{authn.AMROTP}
		channel := i.Claims[AuthenticatorClaimOOBOTPChannelType].(string)
		switch authn.AuthenticatorOOBChannel(channel) {
		case authn.AuthenticatorOOBChannelSMS:
			out = append(out, authn.AMRSMS)
		case authn.AuthenticatorOOBChannelEmail:
			break
		default:
			panic("authenticator: unknown OOB channel: " + channel)
		}
		return out
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
	case authn.AuthenticatorTypePassword:
		// If they are password, they have the same primary/secondary tag.
		return i.Kind == that.Kind
	case authn.AuthenticatorTypeTOTP:
		// If they are TOTP, they have the same secret, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		return subtle.ConstantTimeCompare([]byte(i.Secret), []byte(that.Secret)) == 1
	case authn.AuthenticatorTypeOOBEmail, authn.AuthenticatorTypeOOBSMS:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if i.Kind != that.Kind {
			return false
		}

		iChannel := i.Claims[AuthenticatorClaimOOBOTPChannelType].(string)
		thatChannel := that.Claims[AuthenticatorClaimOOBOTPChannelType].(string)
		if iChannel != thatChannel {
			return false
		}

		switch authn.AuthenticatorOOBChannel(iChannel) {
		case authn.AuthenticatorOOBChannelEmail:
			iEmail := i.Claims[AuthenticatorClaimOOBOTPEmail].(string)
			thatEmail := that.Claims[AuthenticatorClaimOOBOTPEmail].(string)
			return iEmail == thatEmail
		case authn.AuthenticatorOOBChannelSMS:
			// Interesting identifier :)
			iPhone := i.Claims[AuthenticatorClaimOOBOTPPhone].(string)
			thatPhone := that.Claims[AuthenticatorClaimOOBOTPPhone].(string)
			return iPhone == thatPhone
		default:
			panic("authenticator: unknown OOB channel: " + iChannel)
		}
	default:
		panic("authenticator: unknown authenticator type: " + i.Type)
	}
}

func (i *Info) StandardClaims() map[authn.ClaimName]string {
	claims := map[authn.ClaimName]string{}
	switch i.Type {
	case authn.AuthenticatorTypePassword:
		break
	case authn.AuthenticatorTypeTOTP:
		break
	case authn.AuthenticatorTypeOOBEmail, authn.AuthenticatorTypeOOBSMS:
		channel := i.Claims[AuthenticatorClaimOOBOTPChannelType].(string)
		switch authn.AuthenticatorOOBChannel(i.Claims[AuthenticatorClaimOOBOTPChannelType].(string)) {
		case authn.AuthenticatorOOBChannelEmail:
			claims[authn.ClaimEmail] = i.Claims[AuthenticatorClaimOOBOTPEmail].(string)
		case authn.AuthenticatorOOBChannelSMS:
			claims[authn.ClaimPhoneNumber] = i.Claims[AuthenticatorClaimOOBOTPPhone].(string)
		default:
			panic("authenticator: unknown OOB channel: " + channel)
		}
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}
