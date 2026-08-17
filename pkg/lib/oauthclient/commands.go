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
func (c *Commands) DeleteClient(ctx context.Context, clientID string) error {
	if err := c.Store.DeleteClientByClientID(ctx, clientID); err != nil {
		return err
	}

	if !c.hooked {
		c.Database.UseHook(ctx, c)
		c.hooked = true
	}
	c.pendingInvalidations = append(c.pendingInvalidations, clientID)
	return nil
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
