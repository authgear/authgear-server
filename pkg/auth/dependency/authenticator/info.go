package authenticator

import "github.com/authgear/authgear-server/pkg/core/authn"

type Info struct {
	ID            string                  `json:"id"`
	Type          authn.AuthenticatorType `json:"type"`
	Secret        string                  `json:"secret"`
	Props         map[string]interface{}  `json:"props"`
	Authenticator interface{}             `json:"-"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Props: i.Props}
}

func (i *Info) ToRef() Ref {
	return Ref{ID: i.ID, Type: i.Type}
}

func (i *Info) AMR() []string {
	switch i.Type {
	case authn.AuthenticatorTypePassword:
		return []string{"pwd"}
	case authn.AuthenticatorTypeTOTP:
		return []string{"otp"}
	case authn.AuthenticatorTypeOOB:
		out := []string{"otp"}
		channel := i.Props[AuthenticatorPropOOBOTPChannelType].(string)
		switch authn.AuthenticatorOOBChannel(channel) {
		case authn.AuthenticatorOOBChannelSMS:
			out = append(out, "sms")
		case authn.AuthenticatorOOBChannelEmail:
			break
		default:
			panic("authenticator: unexpected OOB channel")
		}
		return out
	case authn.AuthenticatorTypeRecoveryCode:
		return []string{}
	case authn.AuthenticatorTypeBearerToken:
		return []string{}
	default:
		panic("authenticator: unexpected authenticator type")
	}
}
