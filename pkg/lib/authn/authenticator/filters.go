package authenticator

import (
	"fmt"

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

func KeepType(types ...authn.AuthenticatorType) Filter {
	return FilterFunc(func(ai *Info) bool {
		for _, t := range types {
			if ai.Type == t {
				return true
			}
		}
		return false
	})
}

func KeepPrimaryAuthenticatorOfIdentity(ii *identity.Info) Filter {
	return FilterFunc(func(ai *Info) bool {
		if ai.Kind != KindPrimary {
			return false
		}

		for _, typ := range ii.PrimaryAuthenticatorTypes() {
			if ai.Type == typ {
				switch {
				case ii.Type == authn.IdentityTypeLoginID &&
					(ai.Type == authn.AuthenticatorTypeOOBEmail || ai.Type == authn.AuthenticatorTypeOOBSMS):
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

// KeepSecondaryAuthenticatorOfIdentity means only Login ID identity needs MFA.
func KeepSecondaryAuthenticatorOfIdentity(ii *identity.Info) Filter {
	return FilterFunc(func(ai *Info) bool {
		if ai.Kind != KindSecondary {
			return false
		}
		switch ii.Type {
		case authn.IdentityTypeLoginID:
			return true
		case authn.IdentityTypeOAuth:
			return false
		case authn.IdentityTypeAnonymous:
			return false
		case authn.IdentityTypeBiometric:
			return false
		default:
			panic(fmt.Sprintf("authenticator: unexpected identity type: %s", ii.Type))
		}
	})
}
