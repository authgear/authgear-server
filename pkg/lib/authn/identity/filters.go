package identity

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
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

func KeepType(types ...model.IdentityType) Filter {
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
	case model.IdentityTypeLoginID:
		return true
	case model.IdentityTypeOAuth:
		return true
	case model.IdentityTypeAnonymous:
		return false
	case model.IdentityTypeBiometric:
		return false
	case model.IdentityTypePasskey:
		return false
	case model.IdentityTypeSIWE:
		return false
	case model.IdentityTypeLDAP:
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", ii.Type))
	}
}
