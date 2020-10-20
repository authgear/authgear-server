package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserService interface {
	GetRaw(id string) (*user.User, error)
	Count() (uint64, error)
	QueryPage(after, before apimodel.PageCursor, first, last *uint64) ([]apimodel.PageItem, error)
}

type UserFacade struct {
	Users       UserService
	Interaction InteractionService
}

func (f *UserFacade) Get(id string) (*user.User, error) {
	return f.Users.GetRaw(id)
}

func (f *UserFacade) QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error) {
	values, err := f.Users.QueryPage(apimodel.PageCursor(args.After), apimodel.PageCursor(args.Before), args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return graphqlutil.NewPageResult(args, ConvertItems(values), graphqlutil.NewLazy(func() (interface{}, error) {
		return f.Users.Count()
	})), nil
}

func (f *UserFacade) Create(identityDef model.IdentityDef, password string) (*user.User, error) {
	graph, err := f.Interaction.Perform(
		interactionintents.NewIntentSignup(),
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
			return nil, interaction.NewInvariantViolated(
				"PasswordRequired",
				"password is required",
				nil,
			)
		}
	}
	if err != nil {
		return nil, err
	}

	userID, ok := graph.GetNewUserID()
	if !ok {
		return nil, apierrors.NewInternalError("user is not created")
	}
	return f.Users.GetRaw(userID)
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
