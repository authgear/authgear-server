package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AppService interface {
	GetMany(id []string) ([]*model.App, error)
	List(userID string) ([]*model.App, error)
	Create(userID string, id string) error
	UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error
}

type AppLoader struct {
	Apps   AppService
	Authz  AuthzService
	loader *graphqlutil.DataLoader `wire:"-"`
}

func (l *AppLoader) Get(id string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(id)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	if l.loader == nil {
		l.loader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			ids := make([]string, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
			}

			items, err := l.Apps.GetMany(ids)
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

func (l *AppLoader) List(userID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Apps.List(userID)
	})
}

func (l *AppLoader) UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(app.ID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		err := l.Apps.UpdateConfig(app, updateFiles, deleteFiles)
		if err != nil {
			return nil, err
		}

		l.loader.Reset(app.ID)
		return l.Get(app.ID), nil
	})
}

func (l *AppLoader) Create(userID string, id string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		err := l.Apps.Create(userID, id)
		if err != nil {
			return nil, err
		}

		return l.Get(id), nil
	})
}
