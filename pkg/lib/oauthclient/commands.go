package oauthclient

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Commands struct {
	Store    *Store
	Database *appdb.Handle
	Cache    *ClientCache

	// hooked and pendingInvalidations are request-scoped mutable state on an
	// otherwise stateless command struct — safe because Commands is
	// constructed fresh per request by wire, exactly like
	// event.Service.DatabaseHooked/NonBlockingEvents. wire.Struct's "*"
	// enumerates every field regardless of exportedness, so wire:"-" is
	// required here even though these fields are unexported.
	hooked               bool     `wire:"-"`
	pendingInvalidations []string `wire:"-"`
}

// LockForClientCount is re-exported from Store so RegistrationHandler
// depends on one collaborator (via dcr.Commands) rather than reaching into
// pkg/lib/oauthclient directly. See Store.LockForClientCount.
func (c *Commands) LockForClientCount(ctx context.Context, source Source) error {
	return c.Store.LockForClientCount(ctx, source)
}

func (c *Commands) CreateClient(ctx context.Context, options *NewClientOptions) (*Client, error) {
	client := c.Store.NewClient(options)
	if err := c.Store.CreateClient(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
}

// DeleteClient invalidates the resolver cache from DidCommitTx, not inline:
// invalidating before commit would leave the cache empty if the transaction
// then rolled back (a harmless extra miss), but the failure mode that
// actually matters is the opposite one — commit succeeds but invalidation
// never runs, and a deleted client keeps authenticating until the cache TTL
// expires. Running only after the commit is known to have succeeded closes
// that gap; dynamicClientCacheTTL is the bound on the residual window if
// this hook itself never runs (e.g. process crash before DidCommitTx).
//
// It returns the deleted client, read before the delete, so the caller can
// audit-log what was removed — the row is gone afterwards.
func (c *Commands) DeleteClient(ctx context.Context, clientID string) (*Client, error) {
	client, err := c.Store.GetClientByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if err := c.Store.DeleteClientByClientID(ctx, clientID); err != nil {
		return nil, err
	}

	if !c.hooked {
		c.Database.UseHook(ctx, c)
		c.hooked = true
	}
	c.pendingInvalidations = append(c.pendingInvalidations, clientID)
	return client, nil
}

// UpsertCIMDClient writes the row and schedules a resolver-cache
// invalidation from DidCommitTx, for the same reason DeleteClient does:
// invalidating before the commit would leave a harmless extra miss if the
// tx rolled back, while skipping invalidation after a successful commit
// would serve stale metadata -- or worse, a cached NEGATIVE entry -- for up
// to dynamicClientCacheTTL.
//
// The negative entry is the case that actually bites. getClientByClientIDCached
// calls Cache.SetNotFound with a 30s TTL every time a client_id misses in
// Postgres. The very first /oauth2/authorize for a new CIMD client_id does
// exactly that -- the freshness read misses, caches "not found", then this
// upsert creates the row. Without invalidation here, the resolveClient call
// immediately after would read that 30s-old negative entry and reject a
// client that was just successfully resolved.
func (c *Commands) UpsertCIMDClient(ctx context.Context, options *UpsertCIMDClientOptions) (*Client, bool, error) {
	client, created, err := c.Store.UpsertCIMDClient(ctx, options)
	if err != nil {
		return nil, false, err
	}

	if !c.hooked {
		c.Database.UseHook(ctx, c)
		c.hooked = true
	}
	c.pendingInvalidations = append(c.pendingInvalidations, options.ClientID)
	return client, created, nil
}

func (c *Commands) WillCommitTx(ctx context.Context) error {
	return nil
}

func (c *Commands) DidCommitTx(ctx context.Context) {
	pending := c.pendingInvalidations
	c.pendingInvalidations = nil
	for _, clientID := range pending {
		_ = c.Cache.Delete(ctx, clientID)
	}
}
