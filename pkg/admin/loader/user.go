package loader

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserService interface {
	GetManyRaw(id []string) ([]*user.User, error)
	Count() (uint64, error)
	QueryPage(after, before model.PageCursor, first, last *uint64) ([]model.PageItem, error)
}

type UserLoader struct {
	Users  UserService
	loader *graphqlutil.DataLoader `wire:"-"`
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
	values, err := l.Users.QueryPage(model.PageCursor(args.After), model.PageCursor(args.Before), args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return graphqlutil.NewPageResult(args, ConvertItems(values), graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Users.Count()
	})), nil
}
