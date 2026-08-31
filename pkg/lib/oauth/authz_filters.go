package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthorizationFilter interface {
	Keep(ctx context.Context, authz *Authorization) bool
}

type AuthorizationFilterFunc func(ctx context.Context, a *Authorization) bool

func (f AuthorizationFilterFunc) Keep(ctx context.Context, a *Authorization) bool {
	return f(ctx, a)
}

func ApplyAuthorizationFilters(ctx context.Context, authzs []*Authorization, filters ...AuthorizationFilter) (out []*Authorization) {
	for _, authz := range authzs {
		keep := true
		for _, f := range filters {
			if !f.Keep(ctx, authz) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, authz)
		}
	}
	return
}

type KeepThirdPartyAuthorizationFilterClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

// KeepThirdPartyAuthorizationFilter keeps only authorizations belonging to a
// third-party client, whatever its source: a static third_party_app client,
// or a DCR/CIMD-resolved client (which is always
// OAuthClientApplicationTypeDynamicThirdParty, i.e. IsThirdParty()).
//
// There is no static-client-id set. An earlier version built one from
// oauthConfig.Clients (authgear.yaml only), which is why a DCR or CIMD
// client's authorization was silently dropped -- that set was the only
// lookup, not a cache. Resolver.ResolveClient already checks authgear.yaml
// first and returns the static config without touching Redis or Postgres,
// so a static client costs exactly what it cost before: one linear scan of
// oauth.clients. A dynamic client costs one Redis GET, cached for 5
// minutes.
type KeepThirdPartyAuthorizationFilter struct {
	Resolver KeepThirdPartyAuthorizationFilterClientResolver
}

func NewKeepThirdPartyAuthorizationFilter(resolver KeepThirdPartyAuthorizationFilterClientResolver) *KeepThirdPartyAuthorizationFilter {
	return &KeepThirdPartyAuthorizationFilter{Resolver: resolver}
}

func (f *KeepThirdPartyAuthorizationFilter) Keep(ctx context.Context, authz *Authorization) bool {
	client := f.Resolver.ResolveClient(ctx, authz.ClientID)
	if client == nil {
		// An unresolvable client_id: a static client removed from
		// authgear.yaml, a deleted dynamic client, or a CIMD client_id whose
		// domain is no longer allowlisted. Dropped, which preserves today's
		// exact behavior for the removed-static-client case.
		//
		// This does mean a user's grant to a since-deleted third-party
		// client is not listed anywhere -- and therefore cannot be found and
		// revoked through this surface. That gap exists today for removed
		// static clients and is not made worse here; revocation is still
		// available through session revocation and the Admin API by
		// authorization id. Keeping such an authorization and rendering the
		// bare client_id was considered and rejected: it would newly surface
		// removed FIRST-party clients (whose grants are deliberately not
		// shown) as unlabelled entries, a bigger behavior change than the
		// gap it closes.
		return false
	}
	return client.IsThirdParty()
}

var _ AuthorizationFilter = &KeepThirdPartyAuthorizationFilter{}
