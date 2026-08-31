package cimd

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

// NewBucketSpecCIMDLogoPerClient is a fixed, non-configurable bucket, unlike
// the document-fetch buckets: a logo fetch requires a persisted client
// record, and records are only created/refreshed by a document fetch, which
// is already limited per project and per IP and capped by the
// oauth_client_cimd quota. What is NOT already bounded is repeated misses
// for a SINGLE client if cache writes are failing or being evicted -- a
// per-client_id concern, so it gets a per-client_id bucket and nothing
// else. Burst 2/minute is deliberately strict: a legitimate cache miss
// needs exactly one fetch, and the negative cache then suppresses retries
// for ten minutes, so 2/minute already allows a retry that legitimate
// traffic never needs.
func NewBucketSpecCIMDLogoPerClient(clientID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitOAuthCIMDLogoPerClient,
		ratelimit.RateLimitGroupOAuthCIMDLogo,
		&config.RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   2,
		},
		ratelimit.OAuthCIMDLogoPerClient,
		// Hashed, unlike Part 4's IP argument: BucketSpec.Key() joins
		// arguments with ":", and a client_id is a URL containing ":", so
		// passing it raw would let one client_id's key collide with
		// another's namespace -- the same reason redisKeyDynamicClient
		// hashes it.
		crypto.SHA256String(clientID),
	)
}
