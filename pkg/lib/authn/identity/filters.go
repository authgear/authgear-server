package identity

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
)

type Filter interface {
	Keep(ii *Info) bool
}

type FilterFunc func(ii *Info) bool

func (f FilterFunc) Keep(ii *Info) bool {
	return f(ii)
}

func ApplyFilters(iis []*Info, filters ...Filter) (out []*Info) {
	for _, ii := range iis {
		keep := true
		for _, f := range filters {
			if !f.Keep(ii) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, ii)
		}
	}
	return
}

func KeepType(types ...authn.IdentityType) Filter {
	return FilterFunc(func(ii *Info) bool {
		for _, t := range types {
			if ii.Type == t {
				return true
			}
		}
		return false
	})
}

// KeepIdentifiable keeps Login ID identity and OAuth identity.
var KeepIdentifiable FilterFunc = func(ii *Info) bool {
	switch ii.Type {
	case authn.IdentityTypeLoginID:
		return true
	case authn.IdentityTypeOAuth:
		return true
	case authn.IdentityTypeAnonymous:
		return false
	case authn.IdentityTypeBiometric:
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", ii.Type))
	}
}

func OmitID(id string) Filter {
	return FilterFunc(func(ii *Info) bool {
		return ii.ID != id
	})
}
