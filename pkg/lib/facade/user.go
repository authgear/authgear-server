package facade

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserProvider interface {
	Create(userID string) (*user.User, error)
	GetRaw(id string) (*user.User, error)
	Count() (uint64, error)
	QueryPage(listOption user.ListOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error)
	AfterCreate(
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
	) error
}

type UserFacade struct {
	UserProvider
	Coordinator *Coordinator
}

func (u UserFacade) CreateByAdmin(identitySpec *identity.Spec, password string) (*user.User, error) {
	return u.Coordinator.UserCreatebyAdmin(identitySpec, password)
}

func (u UserFacade) Delete(userID string) error {
	return u.Coordinator.UserDelete(userID, false)
}

func (u UserFacade) DeleteFromScheduledDeletion(userID string) error {
	return u.Coordinator.UserDelete(userID, true)
}

func (u UserFacade) Disable(userID string, reason *string) error {
	return u.Coordinator.UserDisable(userID, reason)
}

func (u UserFacade) Reenable(userID string) error {
	return u.Coordinator.UserReenable(userID)
}

func (u UserFacade) ScheduleDeletionByAdmin(userID string) error {
	return u.Coordinator.UserScheduleDeletionByAdmin(userID)
}

func (u UserFacade) UnscheduleDeletionByAdmin(userID string) error {
	return u.Coordinator.UserUnscheduleDeletionByAdmin(userID)
}

func (u UserFacade) ScheduleDeletionByEndUser(userID string) error {
	return u.Coordinator.UserScheduleDeletionByEndUser(userID)
}

func (u UserFacade) Anonymize(userID string) error {
	return u.Coordinator.UserAnonymize(userID, false)
}

func (u UserFacade) AnonymizeFromScheduledAnonymization(userID string) error {
	return u.Coordinator.UserAnonymize(userID, true)
}

func (u UserFacade) ScheduleAnonymizationByAdmin(userID string) error {
	return u.Coordinator.UserScheduleAnonymizationByAdmin(userID)
}

func (u UserFacade) UnscheduleAnonymizationByAdmin(userID string) error {
	return u.Coordinator.UserUnscheduleAnonymizationByAdmin(userID)
}

func (u UserFacade) CheckUserAnonymized(userID string) error {
	return u.Coordinator.UserCheckAnonymized(userID)
}
