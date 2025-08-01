package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ResourceClientLoaderResources interface {
	GetManyResourceClientIDs(ctx context.Context, resourceIDs []string) (map[string][]string, error)
}

type ResourceClientLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Resources ResourceClientLoaderResources
}

func NewResourceClientLoader(resources ResourceClientLoaderResources) *ResourceClientLoader {
	l := &ResourceClientLoader{
		Resources: resources,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *ResourceClientLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	resourceIDs := make([]string, len(keys))
	for i, key := range keys {
		resourceIDs[i] = key.(string)
	}

	// Get entities (map of resourceID to clientIDs).
	resourceIDToClientIDsMap, err := l.Resources.GetManyResourceClientIDs(ctx, resourceIDs)
	if err != nil {
		return nil, err
	}

	out := make([]interface{}, len(resourceIDs))
	for idx, resourceID := range resourceIDs {
		clientIDsForResource, ok := resourceIDToClientIDsMap[resourceID]
		if !ok {
			out[idx] = []string{}
		} else {
			out[idx] = clientIDsForResource
		}
	}
	return out, nil
}
