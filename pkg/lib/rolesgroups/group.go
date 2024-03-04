package rolesgroups

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type NewGroupOptions struct {
	Key         string
	Name        *string
	Description *string
}

type UpdateGroupOptions struct {
	ID             string
	NewKey         *string
	NewName        *string
	NewDescription *string
}

func (o *UpdateGroupOptions) RequireUpdate() bool {
	return o.NewKey != nil || o.NewName != nil || o.NewDescription != nil
}

type Group struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Key         string
	Name        *string
	Description *string
}

func (r *Group) ToModel() *model.Group {
	return &model.Group{
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
