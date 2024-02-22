package rolesgroups

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type NewRoleOptions struct {
	Key         string
	Name        *string
	Description *string
}

type Role struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Key         string
	Name        *string
	Description *string
}

func (r *Role) ToModel() *model.Role {
	return &model.Role{
		Meta: model.Meta{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
	}
}
