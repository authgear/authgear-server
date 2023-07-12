package attrs

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type T struct {
	Pointer string      `json:"pointer"`
	Value   interface{} `json:"value,omitempty"`
}

type List []T

func (l List) AddAbsent(allAbsent []string) List {
	out := make(List, len(l)+len(allAbsent))
	copy(out, l)
	for i, ptr := range allAbsent {
		out[i+len(l)] = T{
			Pointer: ptr,
			Value:   nil,
		}

	}
	return out
}

func (l List) Separate(cfg *config.UserProfileConfig) (stdAttrs List, customAttrs List, unknownAttrs List) {
	customAttrsPointers := make(map[string]struct{})
	for _, c := range cfg.CustomAttributes.Attributes {
		customAttrsPointers[c.Pointer] = struct{}{}
	}

	for _, attr := range l {
		attr := attr
		if IsStandardAttributePointer(attr.Pointer) {
			stdAttrs = append(stdAttrs, attr)
		} else {
			_, ok := customAttrsPointers[attr.Pointer]
			if ok {
				customAttrs = append(customAttrs, attr)
			} else {
				unknownAttrs = append(unknownAttrs, attr)
			}
		}
	}

	return
}

// IsStandardAttributePointer reports if ptrStr is a valid json pointer to a standard attribute.
func IsStandardAttributePointer(ptrStr string) bool {
	switch ptrStr {
	case "/email":
		fallthrough
	case "/phone_number":
		fallthrough
	case "/preferred_username":
		fallthrough
	case "/family_name":
		fallthrough
	case "/given_name":
		fallthrough
	case "/middle_name":
		fallthrough
	case "/name":
		fallthrough
	case "/nickname":
		fallthrough
	case "/picture":
		fallthrough
	case "/profile":
		fallthrough
	case "/website":
		fallthrough
	case "/gender":
		fallthrough
	case "/birthdate":
		fallthrough
	case "/zoneinfo":
		fallthrough
	case "/locale":
		fallthrough
	case "/address":
		fallthrough
	case "/address/formatted":
		fallthrough
	case "/address/street_address":
		fallthrough
	case "/address/locality":
		fallthrough
	case "/address/region":
		fallthrough
	case "/address/postal_code":
		fallthrough
	case "/address/country":
		return true
	default:
		return false
	}
}
