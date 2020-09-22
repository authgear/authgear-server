package loader

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
	GetManyRaw(id []string) ([]*user.User, error)
	Count() (uint64, error)
	QueryPage(after, before apimodel.PageCursor, first, last *uint64) ([]apimodel.PageItem, error)
}

type UserLoader struct {
	Users       UserService
	Interaction InteractionService
	loader      *graphqlutil.DataLoader `wire:"-"`
}

func (l *UserLoader) Get(id string) *graphqlutil.Lazy {
	if l.loader == nil {
		l.loader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			ids := make([]string, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
			}

			users, err := l.Users.GetManyRaw(ids)
			if err != nil {
				return nil, err
			}

			userMap := make(map[string]interface{})
			for _, u := range users {
				userMap[u.ID] = u
			}
			values := make([]interface{}, len(keys))
			for i, id := range ids {
				values[i] = userMap[id]
			}
			return values, nil
		})
	}
	return l.loader.Load(id)
}

func (l *UserLoader) QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error) {
	values, err := l.Users.QueryPage(apimodel.PageCursor(args.After), apimodel.PageCursor(args.Before), args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return graphqlutil.NewPageResult(args, ConvertItems(values), graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Users.Count()
	})), nil
}

func (l *UserLoader) Create(identityDef model.IdentityDef, password string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		var input interface{} = &addIdentityInput{identityDef: identityDef}
		if password != "" {
			input = &addPasswordInput{inner: input, password: password}
		}

		graph, err := l.Interaction.Perform(
			interactionintents.NewIntentSignup(),
			input,
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
		return l.Get(userID), nil
	})
}
