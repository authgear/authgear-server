package authenticator

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type Info struct {
	ID        string                  `json:"id"`
	Labels    map[string]interface{}  `json:"labels"`
	UserID    string                  `json:"user_id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	Type      authn.AuthenticatorType `json:"type"`
	Secret    string                  `json:"secret"`
	Tag       []string                `json:"tag,omitempty"`
	Claims    map[string]interface{}  `json:"claims"`
}

func (i *Info) ToSpec() Spec {
	return Spec{
		UserID: i.UserID,
		Type:   i.Type,
		Tag:    i.Tag,
		Claims: i.Claims,
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
	case authn.AuthenticatorTypeOOB:
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

	iPrimary := slice.ContainsString(i.Tag, TagPrimaryAuthenticator)
	thatPrimary := slice.ContainsString(that.Tag, TagPrimaryAuthenticator)

	iSecondary := slice.ContainsString(i.Tag, TagSecondaryAuthenticator)
	thatSecondary := slice.ContainsString(that.Tag, TagSecondaryAuthenticator)

	switch i.Type {
	case authn.AuthenticatorTypePassword:
		// If they are password, they have the same primary/secondary tag.
		return (iPrimary && thatPrimary) || (iSecondary && thatSecondary)
	case authn.AuthenticatorTypeTOTP:
		// If they are TOTP, they have the same secret, and primary/secondary tag.
		if (iPrimary != thatPrimary) || (iSecondary != thatSecondary) {
			return false
		}

		return subtle.ConstantTimeCompare([]byte(i.Secret), []byte(that.Secret)) == 1
	case authn.AuthenticatorTypeOOB:
		// If they are OOB, they have the same channel, target, and primary/secondary tag.
		if (iPrimary != thatPrimary) || (iSecondary != thatSecondary) {
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
