package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateResource(ctx context.Context, options *NewResourceOptions) (*model.Resource, error) {
	resource := c.Store.NewResource(options)
	err := c.Store.CreateResource(ctx, resource)
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
	resource, err := c.Store.GetResourceByID(ctx, options.ID)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) DeleteResource(ctx context.Context, id string) error {
	return c.Store.DeleteResource(ctx, id)
}

func (c *Commands) AddResourceToClientID(ctx context.Context, resourceURI, clientID string) error {
	// TODO: Implement resource-clientID association by resource URI
	return nil
}

func (c *Commands) RemoveResourceFromClientID(ctx context.Context, resourceURI, clientID string) error {
	// TODO: Implement resource-clientID disassociation by resource URI
	return nil
}

func (c *Commands) CreateScope(ctx context.Context, options *NewScopeOptions) (*model.Scope, error) {
	scope := c.Store.NewScope(options)
	err := c.Store.CreateScope(ctx, scope)
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
	scope, err := c.Store.GetScopeByID(ctx, options.ID)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) DeleteScope(ctx context.Context, id string) error {
	return c.Store.DeleteScope(ctx, id)
}
