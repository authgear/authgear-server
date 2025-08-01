package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Queries struct {
	Store *Store
}

func (q *Queries) GetResourceByID(ctx context.Context, id string) (*model.Resource, error) {
	resource, err := q.Store.GetResourceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (q *Queries) GetResourceByURI(ctx context.Context, uri string) (*model.Resource, error) {
	resource, err := q.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (q *Queries) GetManyResources(ctx context.Context, ids []string) ([]*model.Resource, error) {
	resources, err := q.Store.GetManyResources(ctx, ids)
	if err != nil {
		return nil, err
	}

	resourceModels := make([]*model.Resource, len(resources))
	for i, r := range resources {
		resourceModels[i] = r.ToModel()
	}

	return resourceModels, nil
}

func (q *Queries) ListResources(ctx context.Context, options *ListResourcesOptions, pageArgs graphqlutil.PageArgs) (*ListResourceResult, error) {
	storeResult, err := q.Store.ListResources(ctx, options, pageArgs)
	if err != nil {
		return nil, err
	}

	modelItems := make([]*model.Resource, len(storeResult.Items))
	for i, r := range storeResult.Items {
		modelItems[i] = r.ToModel()
	}

	return &ListResourceResult{
		Items:      modelItems,
		Offset:     storeResult.Offset,
		TotalCount: storeResult.TotalCount,
	}, nil
}

func (q *Queries) GetScope(ctx context.Context, resourceURI string, scope string) (*model.Scope, error) {
	resource, err := q.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}
	sc, err := q.Store.GetResourceScope(ctx, resource.ID, scope)
	if err != nil {
		return nil, err
	}
	return sc.ToModel(), nil
}

func (q *Queries) ListScopes(ctx context.Context, resourceID string, options *ListScopeOptions, pageArgs graphqlutil.PageArgs) (*ListScopeResult, error) {
	storeResult, err := q.Store.ListScopes(ctx, resourceID, options, pageArgs)
	if err != nil {
		return nil, err
	}

	modelItems := make([]*model.Scope, len(storeResult.Items))
	for i, s := range storeResult.Items {
		modelItems[i] = s.ToModel()
	}

	return &ListScopeResult{
		Items:      modelItems,
		Offset:     storeResult.Offset,
		TotalCount: storeResult.TotalCount,
	}, nil
}

func (q *Queries) GetManyScopes(ctx context.Context, ids []string) ([]*model.Scope, error) {
	scopes, err := q.Store.GetManyScopes(ctx, ids)
	if err != nil {
		return nil, err
	}

	scopeModels := make([]*model.Scope, len(scopes))
	for i, s := range scopes {
		scopeModels[i] = s.ToModel()
	}

	return scopeModels, nil
}

func (q *Queries) GetManyResourceClientIDs(ctx context.Context, resourceIDs []string) (map[string][]string, error) {
	return q.Store.ListClientIDsByResourceIDs(ctx, resourceIDs)
}
