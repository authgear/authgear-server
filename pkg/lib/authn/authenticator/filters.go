package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type Filter interface {
	Keep(ai *Info) bool
}

type FilterFunc func(ai *Info) bool

func (f FilterFunc) Keep(ai *Info) bool {
	return f(ai)
}

func ApplyFilters(ais []*Info, filters ...Filter) (out []*Info) {
	for _, a := range ais {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, a)
		}
	}
	return
}

var KeepDefault FilterFunc = func(ai *Info) bool {
	return ai.IsDefault
}

func KeepKind(kind Kind) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Kind == kind
	})
}

func KeepType(types ...model.AuthenticatorType) Filter {
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
				case ii.Type == model.IdentityTypeLoginID &&
					(ai.Type == model.AuthenticatorTypeOOBEmail || ai.Type == model.AuthenticatorTypeOOBSMS):
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

var KeepPrimaryAuthenticatorCanHaveMFA = FilterFunc(func(ai *Info) bool {
	return ai.CanHaveMFA()
})
