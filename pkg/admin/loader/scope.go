package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ScopeLoaderScopes interface {
	GetManyScopes(ctx context.Context, ids []string) ([]*model.Scope, error)
}

type ScopeLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Scopes ScopeLoaderScopes
}

func NewScopeLoader(scopes ScopeLoaderScopes) *ScopeLoader {
	l := &ScopeLoader{
		Scopes: scopes,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *ScopeLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Scopes.GetManyScopes(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Scope)
	for _, entity := range entities {
		entityMap[entity.Meta.ID] = entity
	}

	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		out[i] = entity
	}
	return out, nil
}
