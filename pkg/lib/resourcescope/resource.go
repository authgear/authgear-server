package resourcescope

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type newResourceURI struct {
	Value string
}

func NewResourceURI(ctx context.Context, str string) newResourceURI {
	err := FormatResourceURI{}.CheckFormat(ctx, str)
	if err != nil {
		// This is a programming error because you should always validate the user input before calling NewResourceURI
		panic(fmt.Errorf("invalid resource uri"))
	}
	return newResourceURI{Value: str}
}

type NewResourceOptions struct {
	URI  newResourceURI
	Name *string
	// AccessPolicy is nil when the caller did not specify one, in which case
	// the resource is created with the zero value (no third-party access).
	AccessPolicy *model.AccessPolicy
}

type UpdateResourceOptions struct {
	ResourceURI string
	NewName     *string
	// AccessPolicy is nil when the caller did not specify one, in which case
	// the existing access policy is left unchanged.
	AccessPolicy *model.AccessPolicy
}

type ListResourcesOptions struct {
	SearchKeyword string
	ClientID      string
}

type Resource struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ResourceURI  string
	Name         *string
	AccessPolicy model.AccessPolicy
}

func (r *Resource) ToModel() *model.Resource {
	return &model.Resource{
		Meta: model.Meta{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		ResourceURI:  r.ResourceURI,
		Name:         r.Name,
		AccessPolicy: r.AccessPolicy,
	}
}

type ListResourceResult struct {
	Items      []*model.Resource
	Offset     uint64
	TotalCount uint64
}
