package loader

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type RoleLoaderRoles interface {
	GetManyRoles(ids []string) ([]*model.Role, error)
}

type RoleLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Roles RoleLoaderRoles
}

func NewRoleLoader(roles RoleLoaderRoles) *RoleLoader {
	l := &RoleLoader{
		Roles: roles,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *RoleLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Roles.GetManyRoles(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Role)
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
