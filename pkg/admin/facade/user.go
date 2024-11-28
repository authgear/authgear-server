package facade

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/search"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

type UserService interface {
	CreateByAdmin(ctx context.Context, identitySpec *identity.Spec, password string, generatePassword bool, sendPassword bool, setPasswordExpired bool) (*user.User, error)
	GetRaw(ctx context.Context, id string) (*user.User, error)
	Count(ctx context.Context) (uint64, error)
	QueryPage(ctx context.Context, listOption user.ListOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error)
	Delete(ctx context.Context, userID string) error
	Disable(ctx context.Context, userID string, reason *string) error
	Reenable(ctx context.Context, userID string) error
	ScheduleDeletionByAdmin(ctx context.Context, userID string) error
	UnscheduleDeletionByAdmin(ctx context.Context, userID string) error
	Anonymize(ctx context.Context, userID string) error
	ScheduleAnonymizationByAdmin(ctx context.Context, userID string) error
	UnscheduleAnonymizationByAdmin(ctx context.Context, userID string) error
	CheckUserAnonymized(ctx context.Context, userID string) error
	UpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error
	GetUsersByStandardAttribute(ctx context.Context, attributeName string, attributeValue string) ([]string, error)
	GetUserByLoginID(ctx context.Context, loginIDKey string, loginIDValue string) (string, error)
	GetUserByOAuth(ctx context.Context, oauthProviderAlias string, oauthProviderUserID string) (string, error)
}

type UserSearchService interface {
	QueryUser(
		ctx context.Context,
		searchKeyword string,
		filterOptions user.FilterOptions,
		sortOption user.SortOption,
		pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *search.Stats, error)
}

type UserFacade struct {
	Clock              clock.Clock
	UserSearchService  UserSearchService
	Users              UserService
	LoginIDConfig      *config.LoginIDConfig
	Authenticators     AuthenticatorService
	StandardAttributes StandardAttributesService
	Interaction        InteractionService
}

func (f *UserFacade) ListPage(ctx context.Context, listOption user.ListOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error) {
	values, err := f.Users.QueryPage(ctx, listOption, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return values, graphqlutil.NewPageResult(pageArgs, len(values), graphqlutil.NewLazy(func() (interface{}, error) {
		return f.Users.Count(ctx)
	})), nil
}

func (f *UserFacade) SearchPage(
	ctx context.Context,
	searchKeyword string,
	filterOptions user.FilterOptions,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error) {
	refs, stats, err := f.UserSearchService.QueryUser(ctx, searchKeyword, filterOptions, sortOption, pageArgs)
	if err != nil {
		return nil, nil, err
	}
	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return stats.TotalCount, nil
	})), nil
}

func (f *UserFacade) Create(ctx context.Context, identityDef model.IdentityDef, password string, generatePassword bool, sendPassword bool, setPasswordExpired bool) (userID string, err error) {
	// NOTE: identityDef is assumed to be a login ID since portal only supports login ID
	loginIDInput := identityDef.(*model.IdentityDefLoginID)
	loginIDKeyCofig, ok := f.LoginIDConfig.GetKeyConfig(loginIDInput.Key)
	if !ok {
		return "", api.NewInvariantViolated("InvalidLoginIDKey", "invalid login ID key", nil)
	}

	identitySpec := &identity.Spec{
		Type: identityDef.Type(),
		LoginID: &identity.LoginIDSpec{
			Key:   loginIDInput.Key,
			Type:  loginIDKeyCofig.Type,
			Value: stringutil.NewUserInputString(loginIDInput.Value),
		},
	}

	user, err := f.Users.CreateByAdmin(ctx,
		identitySpec,
		password,
		generatePassword,
		sendPassword,
		setPasswordExpired,
	)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

func (f *UserFacade) ResetPassword(ctx context.Context, id string, password string, generatePassword bool, sendPassword bool, changeOnLogin bool) (err error) {
	err = f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	_, err = f.Interaction.Perform(
		ctx,
		interactionintents.NewIntentResetPassword(),
		&resetPasswordInput{userID: id, password: password, generatePassword: generatePassword, sendPassword: sendPassword, changeOnLogin: changeOnLogin},
	)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) SetPasswordExpired(ctx context.Context, id string, isExpired bool) error {
	err := f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	passwordType := apimodel.AuthenticatorTypePassword
	primaryKind := authenticator.KindPrimary
	ars, err := f.Authenticators.ListRefsByUsers(
		ctx,
		[]string{id},
		&passwordType,
		&primaryKind,
	)
	if err != nil {
		return err
	}

	if len(ars) == 0 {
		return api.ErrAuthenticatorNotFound
	}

	for _, ai := range ars {
		a, err := f.Authenticators.Get(ctx, ai.ID)
		if err != nil {
			return err
		}

		if a.Password == nil {
			continue
		}

		var expireAfter *time.Time
		if isExpired {
			now := f.Clock.NowUTC()
			expireAfter = &now
		}

		_, a, err = f.Authenticators.UpdatePassword(ctx, a, &service.UpdatePasswordOptions{
			SetExpireAfter: true,
			ExpireAfter:    expireAfter,
		})
		if err != nil {
			return err
		}

		err = f.Authenticators.Update(ctx, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *UserFacade) SetDisabled(ctx context.Context, id string, isDisabled bool, reason *string) error {
	var err error
	if isDisabled {
		err = f.Users.Disable(ctx, id, reason)
	} else {
		err = f.Users.Reenable(ctx, id)
	}
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) ScheduleDeletion(ctx context.Context, id string) error {
	err := f.Users.ScheduleDeletionByAdmin(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) UnscheduleDeletion(ctx context.Context, id string) error {
	err := f.Users.UnscheduleDeletionByAdmin(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) Delete(ctx context.Context, id string) error {
	err := f.Users.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) ScheduleAnonymization(ctx context.Context, id string) (err error) {
	err = f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	err = f.Users.ScheduleAnonymizationByAdmin(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) UnscheduleAnonymization(ctx context.Context, id string) (err error) {
	err = f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	err = f.Users.UnscheduleAnonymizationByAdmin(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) Anonymize(ctx context.Context, id string) (err error) {
	err = f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	err = f.Users.Anonymize(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) SetMFAGracePeriod(ctx context.Context, id string, endAt *time.Time) error {
	err := f.Users.CheckUserAnonymized(ctx, id)
	if err != nil {
		return err
	}

	err = f.Users.UpdateMFAEnrollment(ctx, id, endAt)
	if err != nil {
		return err
	}

	return nil
}

func (f *UserFacade) GetUsersByStandardAttribute(ctx context.Context, attributeKey string, attributeValue string) ([]string, error) {
	values, err := f.Users.GetUsersByStandardAttribute(ctx, attributeKey, attributeValue)
	if err != nil {
		return make([]string, 0), err
	}

	return values, nil
}

func (f *UserFacade) GetUserByLoginID(ctx context.Context, loginIDKey string, loginIDValue string) (string, error) {
	value, err := f.Users.GetUserByLoginID(ctx, loginIDKey, loginIDValue)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (f *UserFacade) GetUserByOAuth(ctx context.Context, oauthProviderAlias string, oauthProviderUserID string) (string, error) {
	value, err := f.Users.GetUserByOAuth(ctx, oauthProviderAlias, oauthProviderUserID)
	if err != nil {
		return "", err
	}

	return value, nil
}
