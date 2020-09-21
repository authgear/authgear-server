package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type Filter interface {
	Keep(ai *Info) bool
}

type FilterFunc func(ai *Info) bool

func (f FilterFunc) Keep(ai *Info) bool {
	return f(ai)
}

var KeepDefault FilterFunc = func(ai *Info) bool {
	return ai.IsDefault
}

func KeepKind(kind Kind) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Kind == kind
	})
}

func KeepType(typ authn.AuthenticatorType) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Type == typ
	})
}

func KeepPrimaryAuthenticatorOfIdentity(ii *identity.Info) Filter {
	return FilterFunc(func(ai *Info) bool {
		if ai.Kind != KindPrimary {
			return false
		}

		types := ii.Type.PrimaryAuthenticatorTypes()

		for _, typ := range types {
			if ai.Type == typ {
				switch {
				case ii.Type == authn.IdentityTypeLoginID && ai.Type == authn.AuthenticatorTypeOOB:
					loginID := ii.Claims[identity.IdentityClaimLoginIDValue]
					email, _ := ai.Claims[AuthenticatorClaimOOBOTPEmail].(string)
					phone, _ := ai.Claims[AuthenticatorClaimOOBOTPPhone].(string)
					if loginID == email || loginID == phone {
						return true
					}
				default:
					return true
				}
			}
		}

		return false
	})
}
