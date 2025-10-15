package facade

import (
	"context"
	"time"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserProvider interface {
	Create(ctx context.Context, userID string) (*user.User, error)
	GetRaw(ctx context.Context, id string) (*user.User, error)
	Count(ctx context.Context) (uint64, error)
	QueryPage(ctx context.Context, listOption user.ListOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error)
	AfterCreate(
		ctx context.Context,
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

func (u UserFacade) CreateByAdmin(ctx context.Context, identitySpec *identity.Spec, opts CreatePasswordOptions) (*user.User, error) {
	return u.Coordinator.UserCreatebyAdmin(ctx, identitySpec, opts)
}

func (u UserFacade) Delete(ctx context.Context, userID string) error {
	return u.Coordinator.UserDelete(ctx, userID, false)
}

func (u UserFacade) DeleteFromScheduledDeletion(ctx context.Context, userID string) error {
	return u.Coordinator.UserDelete(ctx, userID, true)
}

func (u UserFacade) Disable(ctx context.Context, options SetDisabledOptions) error {
	return u.Coordinator.UserDisable(ctx, options)
}

func (u UserFacade) Reenable(ctx context.Context, userID string) error {
	return u.Coordinator.UserReenable(ctx, userID)
}

func (u UserFacade) ScheduleDeletionByAdmin(ctx context.Context, userID string) error {
	return u.Coordinator.UserScheduleDeletionByAdmin(ctx, userID)
}

func (u UserFacade) UnscheduleDeletionByAdmin(ctx context.Context, userID string) error {
	return u.Coordinator.UserUnscheduleDeletionByAdmin(ctx, userID)
}

func (u UserFacade) ScheduleDeletionByEndUser(ctx context.Context, userID string) error {
	return u.Coordinator.UserScheduleDeletionByEndUser(ctx, userID)
}

func (u UserFacade) Anonymize(ctx context.Context, userID string) error {
	return u.Coordinator.UserAnonymize(ctx, userID, false)
}

func (u UserFacade) AnonymizeFromScheduledAnonymization(ctx context.Context, userID string) error {
	return u.Coordinator.UserAnonymize(ctx, userID, true)
}

func (u UserFacade) ScheduleAnonymizationByAdmin(ctx context.Context, userID string) error {
	return u.Coordinator.UserScheduleAnonymizationByAdmin(ctx, userID)
}

func (u UserFacade) UnscheduleAnonymizationByAdmin(ctx context.Context, userID string) error {
	return u.Coordinator.UserUnscheduleAnonymizationByAdmin(ctx, userID)
}

func (u UserFacade) CheckUserAnonymized(ctx context.Context, userID string) error {
	return u.Coordinator.UserCheckAnonymized(ctx, userID)
}

func (u UserFacade) UpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error {
	return u.Coordinator.UserUpdateMFAEnrollment(ctx, userID, endAt)
}

func (u UserFacade) GetUsersByStandardAttribute(ctx context.Context, attributeKey string, attributeValue string) ([]string, error) {
	return u.Coordinator.GetUsersByStandardAttribute(ctx, attributeKey, attributeValue)
}

func (u UserFacade) GetUserByLoginID(ctx context.Context, loginIDKey string, loginIDValue string) (string, error) {
	return u.Coordinator.GetUserByLoginID(ctx, loginIDKey, loginIDValue)
}

func (u UserFacade) GetUserByOAuth(ctx context.Context, oauthProviderAlias string, oauthProviderUserID string) (string, error) {
	return u.Coordinator.GetUserByOAuth(ctx, oauthProviderAlias, oauthProviderUserID)
}

func (u UserFacade) GetUserIDsByLoginHint(ctx context.Context, hint *oauth.LoginHint) ([]string, error) {
	return u.Coordinator.GetUserIDsByLoginHint(ctx, hint)
}
