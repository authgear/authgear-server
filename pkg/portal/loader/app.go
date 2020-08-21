package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AppService interface {
	GetManyRaw(id []string) ([]*model.App, error)
	Count() (uint64, error)
	QueryPage(after, before graphqlutil.Cursor, first, last *uint64) ([]graphqlutil.PageItem, error)
}

type AppLoader struct {
	Apps   AppService
	loader *graphqlutil.DataLoader `wire:"-"`
}

func (l *AppLoader) Get(id string) *graphqlutil.Lazy {
	if l.loader == nil {
		l.loader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			ids := make([]string, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
			}

			items, err := l.Apps.GetManyRaw(ids)
			if err != nil {
				return nil, err
			}

			itemMap := make(map[string]interface{})
			for _, u := range items {
				itemMap[u.ID] = u
			}
			values := make([]interface{}, len(keys))
			for i, id := range ids {
				values[i] = itemMap[id]
			}
			return values, nil
		})
	}
	return l.loader.Load(id)
}

func (l *AppLoader) QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error) {
	values, err := l.Apps.QueryPage(args.After, args.Before, args.First, args.Last)
	if err != nil {
		return nil, err
	}

	return graphqlutil.NewPageResult(args, values, graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Apps.Count()
	})), nil
}
