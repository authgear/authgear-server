package authenticator

import (
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/utils"
)

type Filter interface {
	Keep(ai *Info) bool
}

type FilterFunc func(ai *Info) bool

func (f FilterFunc) Keep(ai *Info) bool {
	return f(ai)
}

func KeepTag(tag string) Filter {
	return FilterFunc(func(ai *Info) bool {
		return utils.StringSliceContains(ai.Tag, tag)
	})
}

func KeepType(typ authn.AuthenticatorType) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Type == typ
	})
}
