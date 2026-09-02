package cimd

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// NewBucketSpecCIMDFetchPerProject and NewBucketSpecCIMDFetchPerIP take the
// resolved fetch rate-limits config rather than constructing a
// config.RateLimitConfig literal, so the built-in rates live in
// OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig.SetDefaults()
// and are not duplicated here -- the same shape as
// handler.NewBucketSpecOAuthRegisterPerIP. They live in pkg/lib/cimd, not
// pkg/lib/oauth/handler, because the consumer here is cimd.Service, and
// keeping them beside it avoids pkg/lib/cimd importing pkg/lib/oauth/handler.

func NewBucketSpecCIMDFetchPerProject(fetch *config.OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig) ratelimit.BucketSpec {
	// No args: BucketSpec.IsGlobal is false, so Limiter keys by app id.
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitOAuthClientIDMetadataDocumentFetchPerProject,
		ratelimit.RateLimitGroupOAuthClientIDMetadataDocumentFetch,
		fetch.PerProject,
		ratelimit.OAuthClientIDMetadataDocumentFetchPerProject,
	)
}

func NewBucketSpecCIMDFetchPerIP(fetch *config.OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig, ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitOAuthClientIDMetadataDocumentFetchPerIP,
		ratelimit.RateLimitGroupOAuthClientIDMetadataDocumentFetch,
		fetch.PerIP,
		ratelimit.OAuthClientIDMetadataDocumentFetchPerIP,
		ip,
	)
}

// Limiter is the one ratelimit.Limiter method RateLimiter needs.
type Limiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

// RateLimiter is the ServiceRateLimiter implementation, replacing Part 3's
// no-op binding.
type RateLimiter struct {
	Limiter            Limiter
	RemoteIP           httputil.RemoteIP
	OAuthFeatureConfig *config.OAuthFeatureConfig
}

// CheckFetchAllowed checks both buckets, per-IP first: an abusive single
// caller is stopped without debiting the project's shared bucket, so it
// cannot deny service to the project's own legitimate CIMD clients (same
// convention as RegistrationHandler.Handle). Both buckets are consumed on
// every attempt, successful or not -- the resource being protected is the
// outbound request itself, not a credential-verification attempt.
//
// Takes no host parameter: neither bucket is scoped by fetch target (see
// docs/plans/cimd/2026-08-28-04-rate-limits.md §1.1 for why there is no
// per-(project, host) bucket).
func (r *RateLimiter) CheckFetchAllowed(ctx context.Context) error {
	fetch := r.OAuthFeatureConfig.GetClientIDMetadataDocument().GetRateLimits().GetFetch()
	specs := []ratelimit.BucketSpec{
		NewBucketSpecCIMDFetchPerIP(fetch, string(r.RemoteIP)),
		NewBucketSpecCIMDFetchPerProject(fetch),
	}
	for _, spec := range specs {
		failed, err := r.Limiter.Allow(ctx, spec)
		if err != nil {
			return err
		}
		if err := failed.Error(); err != nil {
			return err
		}
	}
	return nil
}
