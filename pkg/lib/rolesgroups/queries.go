package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Queries struct {
	Store *Store
}

func (q *Queries) GetManyRoles(ids []string) ([]*model.Role, error) {
	roles, err := q.Store.GetManyRoles(ids)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) GetManyGroups(ids []string) ([]*model.Group, error) {
	groups, err := q.Store.GetManyGroups(ids)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListGroupsByRoleID(roleID string) ([]*model.Group, error) {
	groups, err := q.Store.ListGroupsByRoleID(roleID)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}
