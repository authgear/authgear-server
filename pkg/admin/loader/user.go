package loader

import (
	"github.com/authgear/authgear-server/pkg/admin/utils"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
)

type UserService interface {
	GetManyRaw(id []string) ([]*user.User, error)
	Count() (uint64, error)
	QueryPage(after, before model.PageCursor, first, last *uint64) ([]model.PageItem, error)
}

type UserLoader struct {
	Users  UserService
	loader *utils.DataLoader `wire:"-"`
}

func (l *UserLoader) Get(id string) func() (*user.User, error) {
	if l.loader == nil {
		l.loader = utils.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
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
	thunk := l.loader.Load(id)
	return func() (*user.User, error) {
		u, err := thunk()
		if err != nil {
			return nil, err
		}
		return u.(*user.User), nil
	}
}

func (l *UserLoader) QueryPage(args PageArgs) (*PageResult, error) {
	values, err := l.Users.QueryPage(args.After, args.Before, args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return NewPageResult(args, values, l.Users.Count), nil
}
