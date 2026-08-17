package oauthclient

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/tester"
)

type Resolver struct {
	OAuthConfig     *config.OAuthConfig
	TesterEndpoints tester.EndpointsProvider
	Queries         *Queries
}

func (r *Resolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	if clientID == tester.ClientIDTester {
		return tester.NewTesterClient(r.TesterEndpoints.TesterURL().String())
	}

	if client, ok := r.OAuthConfig.GetClient(clientID); ok {
		return client
	}

	if !r.isDynamicClientIDCandidate(clientID) {
		// Fast path: never hits Redis or the DB for a static-shaped
		// client_id. Nothing downstream of this guard runs for a static
		// client or an unknown, non-dynamic-shaped one — no Redis
		// round-trip, no Postgres connection checkout, no BEGIN/COMMIT.
		return nil
	}

	// Redis-first, Postgres only on a miss, and the database scope (when
	// needed) is opened inside Queries rather than here — see
	// docs/plans/dcr/2026-08-17-03-client-resolution.md §3.1, §3.2, §4.2.
	client, err := r.Queries.GetClientConfigByClientID(ctx, clientID)
	if err != nil {
		return nil
	}
	return client
}

// isDynamicClientIDCandidate is a predicate, not an inline strings.HasPrefix
// check, because CIMD adds a second shape to it later: cimd.md's resolution
// order is "Static and DCR clients are always checked first — a string is
// only ever treated as a CIMD candidate once neither matches", so this
// function is the one and only extension point for that.
func (r *Resolver) isDynamicClientIDCandidate(clientID string) bool {
	// DCR: server-generated, always dcrc_-prefixed.
	if IsDCRClientID(clientID) {
		return true
	}
	// CIMD will add: IsCIMDClientIDURL(clientID) && r.OAuthConfig.ClientIDMetadataDocument.IsEnabled()
	// gated on CIMD being enabled, so a URL-shaped client_id costs nothing
	// with CIMD off.
	return false
}
