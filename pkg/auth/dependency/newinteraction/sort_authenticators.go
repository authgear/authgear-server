package newinteraction

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type SortableAuthenticator interface {
	AuthenticatorType() authn.AuthenticatorType
	HasDefaultTag() bool
}

type SortableAuthenticatorInfo authenticator.Info

func (i *SortableAuthenticatorInfo) AuthenticatorType() authn.AuthenticatorType {
	return i.Type
}

func (i *SortableAuthenticatorInfo) HasDefaultTag() bool {
	return slice.ContainsString(i.Tag, authenticator.TagDefaultAuthenticator)
}

// SortAuthenticators sorts slice in-place by considering preferred as the order.
// The item in the slice must somehow associated with a single AuthenticatorType.
func SortAuthenticators(
	preferred []authn.AuthenticatorType,
	slice interface{},
	toSortable func(i int) SortableAuthenticator,
) {
	rank := make(map[authn.AuthenticatorType]int)
	for i, typ := range preferred {
		rank[typ] = i
	}

	sort.SliceStable(slice, func(i, j int) bool {
		iSortable := toSortable(i)
		jSortable := toSortable(j)

		iDefault := iSortable.HasDefaultTag()
		jDefault := jSortable.HasDefaultTag()
		switch {
		case iDefault && !jDefault:
			return true
		case !iDefault && jDefault:
			return false
		default:
			iType := iSortable.AuthenticatorType()
			jType := jSortable.AuthenticatorType()

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
		}

		panic("unreachable")
	})
}
