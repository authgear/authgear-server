package resourcescope

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type NewResourceOptions struct {
	URI  string
	Name *string
}

type UpdateResourceOptions struct {
	ID      string
	NewURI  *string
	NewName *string
}

type Resource struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	URI       string
	Name      *string
}

func (r *Resource) ToModel() *model.Resource {
	return &model.Resource{
		Meta: model.Meta{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		URI:  r.URI,
		Name: r.Name,
	}
}
