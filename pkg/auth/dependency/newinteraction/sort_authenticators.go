package newinteraction

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

// SortAuthenticators sorts ais by considering preferred as the order.
func SortAuthenticators(ais []*authenticator.Info, preferred []authn.AuthenticatorType) []*authenticator.Info {
	rank := make(map[authn.AuthenticatorType]int)
	for i, typ := range preferred {
		rank[typ] = i
	}

	tmp := make([]*authenticator.Info, len(ais))
	copy(tmp, ais)
	ais = tmp

	sort.SliceStable(ais, func(i, j int) bool {
		iRank, iIsPreferred := rank[ais[i].Type]
		jRank, jIsPreferred := rank[ais[j].Type]
		switch {
		case iIsPreferred && jIsPreferred:
			return iRank < jRank
		case !iIsPreferred && !jIsPreferred:
			return false
		case iIsPreferred && !jIsPreferred:
			return true
		case !iIsPreferred && jIsPreferred:
			return false
		}
		panic("unreachable")
	})

	return ais
}
