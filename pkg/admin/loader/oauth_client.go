package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DynamicClientLoaderClients interface {
	GetManyClientModels(ctx context.Context, ids []string) ([]*model.OAuthClient, error)
}

type DynamicClientLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Clients DynamicClientLoaderClients
}

func NewDynamicClientLoader(clients DynamicClientLoaderClients) *DynamicClientLoader {
	l := &DynamicClientLoader{
		Clients: clients,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *DynamicClientLoader) LoadFunc(ctx context.Context, keys []any) ([]any, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Clients.GetManyClientModels(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.OAuthClient)
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	out := make([]any, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		out[i] = entity
	}
	return out, nil
}
