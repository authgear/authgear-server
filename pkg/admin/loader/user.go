package loader

import "github.com/authgear/authgear-server/pkg/lib/api/model"

type UserService interface {
	Get(id string) (*model.User, error)
	Count() (uint64, error)
	QueryPage(after, before model.PageCursor, first, last *uint64) ([]model.PageItem, error)
}

type UserLoader struct {
	Users UserService
}

func (l *UserLoader) Get(id string) (interface{}, error) {
	return l.Users.Get(id)
}

func (l *UserLoader) QueryPage(args PageArgs) (*PageResult, error) {
	values, err := l.Users.QueryPage(args.After, args.Before, args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return NewPageResult(args, values, l.Users.Count), nil
}
