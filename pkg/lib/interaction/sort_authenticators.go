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
	// Non-passkey authenticators must come BEFORE passkey authenticators.
	orderByPasskey := func(i, j int) bool {
		iSortable := toSortable(i)
		jSortable := toSortable(j)

		iType := iSortable.AuthenticatorType()
		jType := jSortable.AuthenticatorType()

		iPasskey := iType == model.AuthenticatorTypePasskey
		jPasskey := jType == model.AuthenticatorTypePasskey

		return !iPasskey && jPasskey
	}

	// Default authenticators must come BEFORE non-default authenticators.
	orderByDefault := func(i, j int) bool {
		iSortable := toSortable(i)
		jSortable := toSortable(j)

		iDefault := iSortable.IsDefaultAuthenticator()
		jDefault := jSortable.IsDefaultAuthenticator()

		return iDefault && !jDefault
	}

	// authenticators with a higher rank (lower rank value) must come BEFORE
	// authenticators with a lower rank (higher rank value).
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
		case iIsPreferred && !jIsPreferred:
			return true
		default:
			return false
		}
	}

	less := sortutil.LessFunc(orderByPasskey).AndThen(orderByDefault).AndThen(orderByRank)
	sort.SliceStable(slice, less)
}
