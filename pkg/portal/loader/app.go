package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AppLoaderAppService interface {
	GetMany(ctx context.Context, ids []string) ([]*model.App, error)
}
type AppLoader struct {
	*graphqlutil.DataLoader `wire:"-"`
	AppService              AppLoaderAppService
	Authz                   AuthzService
}

func NewAppLoader(appService AppLoaderAppService, authz AuthzService) *AppLoader {
	l := &AppLoader{
		AppService: appService,
		Authz:      authz,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *AppLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	apps, err := l.AppService.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.App)
	for _, app := range apps {
		entityMap[app.ID] = app
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		_, err := l.Authz.CheckAccessOfViewer(ctx, entity.ID)
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
