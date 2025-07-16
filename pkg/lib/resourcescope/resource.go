package resourcescope

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type NewResourceOptions struct {
	URI  string
	Name *string
}

type UpdateResourceOptions struct {
	ResourceURI string
	NewName     *string
}

type ListResourcesOptions struct {
	SearchKeyword string
	ClientID      string
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

type ListResourceResult struct {
	Items      []*model.Resource
	Offset     uint64
	TotalCount uint64
}

var uriSchema = validation.NewSimpleSchema(`
	{
		"type": "string",
		"minLength": 1,
		"maxLength": 100,
		"format": "x_resource_uri"
	}
`)

func ValidateResourceURI(ctx context.Context, uri string) error {
	return uriSchema.Validator().ValidateValue(ctx, uri)
}
