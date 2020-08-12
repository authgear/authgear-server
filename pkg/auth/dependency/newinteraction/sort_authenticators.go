package newinteraction

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/core/authn"
)

// SortAuthenticators sorts slice in-place by considering preferred as the order.
// The item in the slice must somehow associated with a single AuthenticatorType.
func SortAuthenticators(
	preferred []authn.AuthenticatorType,
	slice interface{},
	getType func(i int) authn.AuthenticatorType,
) {
	rank := make(map[authn.AuthenticatorType]int)
	for i, typ := range preferred {
		rank[typ] = i
	}

	sort.SliceStable(slice, func(i, j int) bool {
		iType := getType(i)
		jType := getType(j)

		iRank, iIsPreferred := rank[iType]
		jRank, jIsPreferred := rank[jType]

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
}
