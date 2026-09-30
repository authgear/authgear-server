package loader

import (
	"context"
	"slices"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

// ResourceScopeKey names a scope within its resource, since a scope name is
// unique only within one resource.
type ResourceScopeKey struct {
	ResourceID string
	Scope      string
}

type ResourceScopeLoaderScopes interface {
	ListScopesByResourceIDs(ctx context.Context, resourceIDs []string, scopes []string) ([]*model.Scope, error)
}

type ResourceScopeLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Scopes ResourceScopeLoaderScopes
}

func NewResourceScopeLoader(scopes ResourceScopeLoaderScopes) *ResourceScopeLoader {
	l := &ResourceScopeLoader{
		Scopes: scopes,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *ResourceScopeLoader) LoadFunc(ctx context.Context, keys []any) ([]any, error) {
	var resourceIDs []string
	var names []string
	for _, key := range keys {
		k := key.(ResourceScopeKey)
		if !slices.Contains(resourceIDs, k.ResourceID) {
			resourceIDs = append(resourceIDs, k.ResourceID)
		}
		if !slices.Contains(names, k.Scope) {
			names = append(names, k.Scope)
		}
	}

	entities, err := l.Scopes.ListScopesByResourceIDs(ctx, resourceIDs, names)
	if err != nil {
		return nil, err
	}

	entityMap := make(map[ResourceScopeKey]*model.Scope)
	for _, entity := range entities {
		entityMap[ResourceScopeKey{ResourceID: entity.ResourceID, Scope: entity.Scope}] = entity
	}

	out := make([]any, len(keys))
	for i, key := range keys {
		out[i] = entityMap[key.(ResourceScopeKey)]
	}
	return out, nil
}
