package facade

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserProvider interface {
	Create(userID string) (*user.User, error)
	GetRaw(id string) (*user.User, error)
	Count() (uint64, error)
	QueryPage(sortOption user.SortOption, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error)
	UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error
}

type UserFacade struct {
	UserProvider
	Coordinator *Coordinator
}

func (u UserFacade) Delete(userID string) error {
	return u.Coordinator.UserDelete(userID)
}
