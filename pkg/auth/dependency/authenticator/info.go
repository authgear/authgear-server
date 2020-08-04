package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Info struct {
	ID     string                  `json:"id"`
	UserID string                  `json:"user_id"`
	Type   authn.AuthenticatorType `json:"type"`
	Secret string                  `json:"secret"`
	Tag    []string                `json:"tag,omitempty"`
	Props  map[string]interface{}  `json:"props"`
}

func (i *Info) ToSpec() Spec {
	return Spec{
		UserID: i.UserID,
		Type:   i.Type,
		Tag:    i.Tag,
		Props:  i.Props,
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
		channel := i.Props[AuthenticatorPropOOBOTPChannelType].(string)
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
