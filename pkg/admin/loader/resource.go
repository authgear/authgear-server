package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ResourceLoaderResources interface {
	GetManyResources(ctx context.Context, ids []string) ([]*model.Resource, error)
}

type ResourceLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Resources ResourceLoaderResources
}

func NewResourceLoader(resources ResourceLoaderResources) *ResourceLoader {
	l := &ResourceLoader{
		Resources: resources,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *ResourceLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Resources.GetManyResources(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Resource)
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		out[i] = entity
	}
	return out, nil
}
