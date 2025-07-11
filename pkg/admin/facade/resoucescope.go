package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ResourceScopeCommands interface {
	CreateResource(ctx context.Context, options *resourcescope.NewResourceOptions) (*model.Resource, error)
	UpdateResource(ctx context.Context, options *resourcescope.UpdateResourceOptions) (*model.Resource, error)
	DeleteResourceByURI(ctx context.Context, uri string) error
	CreateScope(ctx context.Context, options *resourcescope.NewScopeOptions) (*model.Scope, error)
	UpdateScope(ctx context.Context, options *resourcescope.UpdateScopeOptions) (*model.Scope, error)
	DeleteScope(ctx context.Context, resourceURI string, scope string) error
	AddResourceToClientID(ctx context.Context, resourceURI string, clientID string) error
	RemoveResourceFromClientID(ctx context.Context, resourceURI string, clientID string) error
}

type ResourceScopeQueries interface {
	GetResourceByURI(ctx context.Context, uri string) (*model.Resource, error)
	GetScope(ctx context.Context, resourceURI string, scope string) (*model.Scope, error)
	ListScopes(ctx context.Context, resourceID string) ([]*model.Scope, error)
	ListResources(ctx context.Context, options *resourcescope.ListResourcesOptions, pageArgs graphqlutil.PageArgs) (*resourcescope.ListResourceResult, error)
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

func (f *ResourceScopeFacade) DeleteResourceByURI(ctx context.Context, uri string) error {
	return f.ResourceScopeCommands.DeleteResourceByURI(ctx, uri)
}

func (f *ResourceScopeFacade) GetResourceByURI(ctx context.Context, uri string) (*model.Resource, error) {
	return f.ResourceScopeQueries.GetResourceByURI(ctx, uri)
}

func (f *ResourceScopeFacade) CreateScope(ctx context.Context, options *resourcescope.NewScopeOptions) (*model.Scope, error) {
	return f.ResourceScopeCommands.CreateScope(ctx, options)
}

func (f *ResourceScopeFacade) UpdateScope(ctx context.Context, options *resourcescope.UpdateScopeOptions) (*model.Scope, error) {
	return f.ResourceScopeCommands.UpdateScope(ctx, options)
}

func (f *ResourceScopeFacade) DeleteScope(ctx context.Context, resourceURI string, scope string) error {
	return f.ResourceScopeCommands.DeleteScope(ctx, resourceURI, scope)
}

func (f *ResourceScopeFacade) GetScope(ctx context.Context, resourceURI string, scope string) (*model.Scope, error) {
	return f.ResourceScopeQueries.GetScope(ctx, resourceURI, scope)
}

func (f *ResourceScopeFacade) ListScopes(ctx context.Context, resourceID string) ([]*model.Scope, error) {
	return f.ResourceScopeQueries.ListScopes(ctx, resourceID)
}

func (f *ResourceScopeFacade) ListResources(ctx context.Context, options *resourcescope.ListResourcesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	result, err := f.ResourceScopeQueries.ListResources(ctx, options, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	refs := make([]model.PageItemRef, len(result.Items))
	for i, r := range result.Items {
		i_uint64 := uint64(i) // #nosec G115
		pageKey := db.PageKey{Offset: result.Offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, nil, err
		}
		refs[i] = model.PageItemRef{ID: r.ID, Cursor: cursor}
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return result.TotalCount, nil
	})), nil
}

func (f *ResourceScopeFacade) AddResourceToClientID(ctx context.Context, resourceURI string, clientID string) error {
	return f.ResourceScopeCommands.AddResourceToClientID(ctx, resourceURI, clientID)
}

func (f *ResourceScopeFacade) RemoveResourceFromClientID(ctx context.Context, resourceURI string, clientID string) error {
	return f.ResourceScopeCommands.RemoveResourceFromClientID(ctx, resourceURI, clientID)
}
