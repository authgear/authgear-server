package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateResource(ctx context.Context, options *NewResourceOptions) (*model.Resource, error) {
	// TODO: implement
	return nil, nil
}

func (c *Commands) UpdateResource(ctx context.Context, options *UpdateResourceOptions) (*model.Resource, error) {
	// TODO: implement
	return nil, nil
}

func (c *Commands) DeleteResource(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (c *Commands) CreateScope(ctx context.Context, options *NewScopeOptions) (*model.Scope, error) {
	// TODO: implement
	return nil, nil
}

func (c *Commands) UpdateScope(ctx context.Context, options *UpdateScopeOptions) (*model.Scope, error) {
	// TODO: implement
	return nil, nil
}

func (c *Commands) DeleteScope(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}
