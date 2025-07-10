package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Queries struct {
	Store *Store
}

func (q *Queries) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	resource, err := q.Store.GetResourceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (q *Queries) ListResources(ctx context.Context) ([]*model.Resource, error) {
	// TODO: implement
	return nil, nil
}

func (q *Queries) GetScope(ctx context.Context, id string) (*model.Scope, error) {
	// TODO: implement
	return nil, nil
}

func (q *Queries) ListScopes(ctx context.Context, resourceID string) ([]*model.Scope, error) {
	// TODO: implement
	return nil, nil
}
