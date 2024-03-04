package loader

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type GroupLoaderGroups interface {
	GetManyGroups(ids []string) ([]*model.Group, error)
}

type GroupLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Groups GroupLoaderGroups
}

func NewGroupLoader(groups GroupLoaderGroups) *GroupLoader {
	l := &GroupLoader{
		Groups: groups,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *GroupLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Groups.GetManyGroups(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Group)
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
