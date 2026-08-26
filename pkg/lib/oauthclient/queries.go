package oauthclient

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Queries struct {
	Store       *Store
	OAuthConfig *config.OAuthConfig
	Database    *appdb.Handle
	Cache       *ClientCache
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

// GetClientConfigByClientID resolves clientID to a synthesized
// *config.OAuthClientConfig for runtime use — see
// docs/plans/dcr/2026-08-17-03-client-resolution.md. Token lifetimes are
// resolved via ResolveTokenLifetimes against the fetched row's own Source,
// exactly like GetClientModelByID/ListClients/GetManyClientModels — not a
// caller-supplied value, since the caller cannot know which config key
// governs the row until after it is fetched. This also means the config
// itself, not the cached row, is what is re-read on every call, so an
// admin edit to default_client_config takes effect immediately.
func (q *Queries) GetClientConfigByClientID(ctx context.Context, clientID string) (*config.OAuthClientConfig, error) {
	c, err := q.getClientByClientIDCached(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return c.ToClientConfig(ResolveTokenLifetimes(q.OAuthConfig, c.Source)), nil
}

// getClientByClientIDCached consults Redis first and only touches Postgres
// on a miss, opening a ReadOnly scope itself when the caller has none —
// ResolveClient is called from middleware, view models and handlers alike,
// many with no active transaction (see the plan's §4.2). On a cache hit
// this opens no database transaction and takes no connection from the
// Postgres pool.
func (q *Queries) getClientByClientIDCached(ctx context.Context, clientID string) (*Client, error) {
	c, found, err := q.Cache.Get(ctx, clientID)
	if err != nil {
		// A cache failure must never take the endpoint down; fall through to
		// the database.
		c, found = nil, false
	}
	if found {
		if c == nil {
			// A cached negative result. Unlike an upstream draft of this
			// plan, this returns ErrDynamicClientNotFound rather than
			// (nil, nil) — the latter would panic downstream the moment a
			// caller like GetClientConfigByClientID dereferences a nil
			// *Client.
			return nil, ErrDynamicClientNotFound
		}
		return c, nil
	}

	read := func(ctx context.Context) error {
		c, err = q.Store.GetClientByClientID(ctx, clientID)
		return err
	}
	if q.Database.IsInTx(ctx) {
		err = read(ctx)
	} else {
		err = q.Database.ReadOnly(ctx, read)
	}

	switch {
	case errors.Is(err, ErrDynamicClientNotFound):
		_ = q.Cache.SetNotFound(ctx, clientID)
		return nil, err
	case err != nil:
		return nil, err
	}
	_ = q.Cache.Set(ctx, c)
	return c, nil
}
