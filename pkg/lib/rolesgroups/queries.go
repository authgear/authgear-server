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
