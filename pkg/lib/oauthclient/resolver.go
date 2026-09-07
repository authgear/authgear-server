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
// check, because CIMD adds a second shape to it: cimd.md's resolution order
// is "Static and DCR clients are always checked first — a string is only
// ever treated as a CIMD candidate once neither matches", so this function
// is the one and only extension point for that.
//
// It answers only "is a persisted-row lookup warranted for this string?". It
// applies no policy at all -- neither trust policy (allowed_domains,
// insecure_http_allowed) nor the `enabled` feature switch gates reading,
// only FETCHING (EnsureClientResolved) does -- see
// docs/plans/cimd/2026-08-28-01-config-and-client-id.md §4.1 for why:
// removing a domain from allowed_domains, or turning CIMD off entirely,
// must not break a client that already resolved successfully. A URL-shaped
// client_id therefore always costs one Queries lookup (Redis, Postgres only
// on a miss), the same as a dcrc_-prefixed one, regardless of `enabled`.
func (r *Resolver) isDynamicClientIDCandidate(clientID string) bool {
	// DCR: server-generated, always dcrc_-prefixed.
	if IsDCRClientID(clientID) {
		return true
	}
	// CIMD: shape only, unconditionally -- see the func comment.
	return IsCIMDClientID(clientID, true)
}
