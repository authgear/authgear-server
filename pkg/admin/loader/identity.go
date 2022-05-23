package loader

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type IdentityLoaderIdentityService interface {
	GetMany(ids []string) ([]*identity.Info, error)
}

type IdentityLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Identities IdentityLoaderIdentityService
}

func NewIdentityLoader(identities IdentityLoaderIdentityService) *IdentityLoader {
	l := &IdentityLoader{
		Identities: identities,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *IdentityLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Identities.GetMany(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*identity.Info)
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, key := range keys {
		entity := entityMap[key.(string)]
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
