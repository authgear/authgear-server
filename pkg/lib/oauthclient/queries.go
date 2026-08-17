package oauthclient

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Queries struct {
	Store       *Store
	OAuthConfig *config.OAuthConfig
}

type ListClientResult struct {
	Items      []*model.OAuthClient
	Offset     uint64
	TotalCount uint64
}

func (q *Queries) GetClientModelByID(ctx context.Context, id string) (*model.OAuthClient, error) {
	c, err := q.Store.GetClientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.ToModel(ResolveTokenLifetimes(q.OAuthConfig, c.Source)), nil
}

func (q *Queries) GetManyClientModels(ctx context.Context, ids []string) ([]*model.OAuthClient, error) {
	cs, err := q.Store.GetManyClientsByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	models := make([]*model.OAuthClient, len(cs))
	for i, c := range cs {
		models[i] = c.ToModel(ResolveTokenLifetimes(q.OAuthConfig, c.Source))
	}
	return models, nil
}

func (q *Queries) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) (*ListClientResult, error) {
	storeResult, err := q.Store.ListClients(ctx, pageArgs)
	if err != nil {
		return nil, err
	}

	modelItems := make([]*model.OAuthClient, len(storeResult.Items))
	for i, c := range storeResult.Items {
		modelItems[i] = c.ToModel(ResolveTokenLifetimes(q.OAuthConfig, c.Source))
	}

	return &ListClientResult{
		Items:      modelItems,
		Offset:     storeResult.Offset,
		TotalCount: storeResult.TotalCount,
	}, nil
}

func (q *Queries) CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error) {
	return q.Store.CountClientsBySource(ctx, source)
}
