package interaction

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/sortutil"
)

type SortableAuthenticator interface {
	AuthenticatorType() model.AuthenticatorType
	IsDefaultAuthenticator() bool
}

type SortableAuthenticatorInfo authenticator.Info

func (i *SortableAuthenticatorInfo) AuthenticatorType() model.AuthenticatorType {
	return i.Type
}

func (i *SortableAuthenticatorInfo) IsDefaultAuthenticator() bool {
	return i.IsDefault
}

// SortAuthenticators sorts slice in-place by considering preferred as the order.
// The item in the slice must somehow associated with a single AuthenticatorType.
func SortAuthenticators(
	preferred []model.AuthenticatorType,
	slice interface{},
	toSortable func(i int) SortableAuthenticator,
) {
	orderByDefault := func(i, j int) bool {
		iSortable := toSortable(i)
		jSortable := toSortable(j)

		iDefault := iSortable.IsDefaultAuthenticator()
		jDefault := jSortable.IsDefaultAuthenticator()

		return iDefault && !jDefault
	}

	rank := make(map[model.AuthenticatorType]int)
	for i, typ := range preferred {
		rank[typ] = i
	}
	orderByRank := func(i, j int) bool {
		iSortable := toSortable(i)
		jSortable := toSortable(j)

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
		panic("unreachable")
	}

	less := sortutil.LessFunc(orderByDefault).AndThen(orderByRank)
	sort.SliceStable(slice, less)
}
