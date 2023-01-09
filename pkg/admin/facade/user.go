package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserService interface {
	GetRaw(id string) (*user.User, error)
	Count() (uint64, error)
	QueryPage(sortOption user.SortOption, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error)
	Delete(userID string) error
	Disable(userID string, reason *string) error
	Reenable(userID string) error
	ScheduleDeletionByAdmin(userID string) error
	UnscheduleDeletionByAdmin(userID string) error
	Anonymize(userID string) error
	ScheduleAnonymizationByAdmin(userID string) error
	UnscheduleAnonymizationByAdmin(userID string) error
}

type UserSearchService interface {
	QueryUser(searchKeyword string, sortOption user.SortOption, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *libes.Stats, error)
}

type UserFacade struct {
	UserSearchService  UserSearchService
	Users              UserService
	StandardAttributes StandardAttributesService
	Interaction        InteractionService
}

func (f *UserFacade) ListPage(sortOption user.SortOption, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error) {
	values, err := f.Users.QueryPage(sortOption, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return values, graphqlutil.NewPageResult(pageArgs, len(values), graphqlutil.NewLazy(func() (interface{}, error) {
		return f.Users.Count()
	})), nil
}

func (f *UserFacade) SearchPage(searchKeyword string, sortOption user.SortOption, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error) {
	refs, stats, err := f.UserSearchService.QueryUser(searchKeyword, sortOption, pageArgs)
	if err != nil {
		return nil, nil, err
	}
	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return stats.TotalCount, nil
	})), nil
}

func (f *UserFacade) Create(identityDef model.IdentityDef, password string) (string, error) {
	graph, err := f.Interaction.Perform(
		&interactionintents.IntentAuthenticate{
			Kind:                     interactionintents.IntentAuthenticateKindSignup,
			SuppressIDPSessionCookie: true,
		},
		&createUserInput{
			identityDef: identityDef,
			password:    password,
		},
	)
	var errInputRequired *interaction.ErrInputRequired
	if errors.As(err, &errInputRequired) {
		switch graph.CurrentNode().(type) {
		case *nodes.NodeCreateAuthenticatorBegin:
			// TODO(interaction): better interpretation of input required error?
			return "", interaction.NewInvariantViolated(
				"PasswordRequired",
				"password is required",
				nil,
			)
		}
	}
	if err != nil {
		return "", err
	}

	userID, ok := graph.GetNewUserID()
	if !ok {
		return "", apierrors.NewInternalError("user is not created")
	}
	return userID, nil
}

func (f *UserFacade) ResetPassword(id string, password string) error {
	_, err := f.Interaction.Perform(
		interactionintents.NewIntentResetPassword(),
		&resetPasswordInput{userID: id, password: password},
	)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) SetDisabled(id string, isDisabled bool, reason *string) error {
	var err error
	if isDisabled {
		err = f.Users.Disable(id, reason)
	} else {
		err = f.Users.Reenable(id)
	}
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) ScheduleDeletion(id string) error {
	err := f.Users.ScheduleDeletionByAdmin(id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) UnscheduleDeletion(id string) error {
	err := f.Users.UnscheduleDeletionByAdmin(id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) Delete(id string) error {
	err := f.Users.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) ScheduleAnonymization(id string) error {
	err := f.Users.ScheduleAnonymizationByAdmin(id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) UnscheduleAnonymization(id string) error {
	err := f.Users.UnscheduleAnonymizationByAdmin(id)
	if err != nil {
		return err
	}
	return nil
}

func (f *UserFacade) Anonymize(id string) error {
	err := f.Users.Anonymize(id)
	if err != nil {
		return err
	}
	return nil
}
