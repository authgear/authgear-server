package authenticator

import (
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/utils"
)

func KeepTag(tag string) func(*Info) bool {
	return func(ai *Info) bool {
		return utils.StringSliceContains(ai.Tag, tag)
	}
}

func KeepType(typ authn.AuthenticatorType) func(*Info) bool {
	return func(ai *Info) bool {
		return ai.Type == typ
	}
}
