package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateResource(ctx context.Context, options *NewResourceOptions) (*model.Resource, error) {
	err := ValidateResourceURI(ctx, options.URI)
	if err != nil {
		return nil, err
	}

	resource := c.Store.NewResource(options)
	err = c.Store.CreateResource(ctx, resource)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) UpdateResource(ctx context.Context, options *UpdateResourceOptions) (*model.Resource, error) {
	err := c.Store.UpdateResource(ctx, options)
	if err != nil {
		return nil, err
	}
	resource, err := c.Store.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) DeleteResourceByURI(ctx context.Context, uri string) error {
	return c.Store.DeleteResourceByURI(ctx, uri)
}

func (c *Commands) GetResourceByURI(ctx context.Context, uri string) (*model.Resource, error) {
	resource, err := c.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) CreateScope(ctx context.Context, options *NewScopeOptions) (*model.Scope, error) {
	err := ValidateScope(ctx, options.Scope)
	if err != nil {
		return nil, err
	}
	resource, err := c.Store.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return nil, err
	}
	scope := c.Store.NewScope(resource, options)
	err = c.Store.CreateScope(ctx, scope)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) UpdateScope(ctx context.Context, options *UpdateScopeOptions) (*model.Scope, error) {
	err := c.Store.UpdateScope(ctx, options)
	if err != nil {
		return nil, err
	}
	scope, err := c.Store.GetScope(ctx, options.ResourceURI, options.Scope)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) DeleteScope(ctx context.Context, resourceURI string, scope string) error {
	return c.Store.DeleteScope(ctx, resourceURI, scope)
}
