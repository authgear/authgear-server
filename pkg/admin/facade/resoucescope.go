package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
)

type ResourceScopeCommands interface {
	CreateResource(ctx context.Context, options *resourcescope.NewResourceOptions) (*model.Resource, error)
	UpdateResource(ctx context.Context, options *resourcescope.UpdateResourceOptions) (*model.Resource, error)
	DeleteResource(ctx context.Context, id string) error
}

type ResourceScopeQueries interface {
	GetResource(ctx context.Context, id string) (*model.Resource, error)
}

type ResourceScopeFacade struct {
	ResourceScopeCommands ResourceScopeCommands
	ResourceScopeQueries  ResourceScopeQueries
}

func (f *ResourceScopeFacade) CreateResource(ctx context.Context, options *resourcescope.NewResourceOptions) (*model.Resource, error) {
	return f.ResourceScopeCommands.CreateResource(ctx, options)
}

func (f *ResourceScopeFacade) UpdateResource(ctx context.Context, options *resourcescope.UpdateResourceOptions) (*model.Resource, error) {
	return f.ResourceScopeCommands.UpdateResource(ctx, options)
}

func (f *ResourceScopeFacade) DeleteResource(ctx context.Context, id string) error {
	return f.ResourceScopeCommands.DeleteResource(ctx, id)
}

func (f *ResourceScopeFacade) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	return f.ResourceScopeQueries.GetResource(ctx, id)
}
