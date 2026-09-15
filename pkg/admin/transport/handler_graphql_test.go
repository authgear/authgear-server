package transport

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type fakeMutationRateLimiter struct {
	calls    []ratelimit.BucketSpec
	ns       []int
	rejectAt int
}

func (l *fakeMutationRateLimiter) AllowN(ctx context.Context, spec ratelimit.BucketSpec, n int) (*ratelimit.FailedReservation, error) {
	l.calls = append(l.calls, spec)
	l.ns = append(l.ns, n)
	if l.rejectAt > 0 && len(l.calls) == l.rejectAt {
		return ratelimit.NewFailedReservation(spec), nil
	}
	return nil, nil
}

func newEnabledScope() *config.AdminAPIRateLimitsMutationScopeFeatureConfig {
	enabled := true
	return &config.AdminAPIRateLimitsMutationScopeFeatureConfig{
		PerProject: &config.RateLimitConfig{Enabled: &enabled, Period: "1m", Burst: 1000},
	}
}

func newHandler(limiter *fakeMutationRateLimiter, scope *config.AdminAPIRateLimitsMutationScopeFeatureConfig) *GraphQLHandler {
	return &GraphQLHandler{
		RateLimiter: limiter,
		AdminAPIFeatureConfig: &config.AdminAPIFeatureConfig{
			RateLimits: &config.AdminAPIRateLimitsFeatureConfig{
				Mutation: &config.AdminAPIRateLimitsMutationFeatureConfig{
					All: scope,
				},
			},
		},
	}
}

func TestCheckMutationRateLimit(t *testing.T) {
	Convey("GraphQLHandler.checkMutationRateLimit", t, func() {
		ctx := context.Background()

		Convey("does not touch the limiter for a query", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 0), ShouldBeNil)
			So(limiter.calls, ShouldBeEmpty)
		})

		Convey("charges one token per mutation field", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 3), ShouldBeNil)

			So(limiter.ns, ShouldResemble, []int{3})
			So(limiter.calls, ShouldHaveLength, 1)
		})

		Convey("charges the per-project bucket, keyed by app rather than caller", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 1), ShouldBeNil)

			So(limiter.calls[0].RateLimitName, ShouldEqual, ratelimit.RateLimitAdminAPIMutationAllPerProject)
			// No bucket arguments: Limiter keys on the app id alone.
			So(limiter.calls[0].Arguments, ShouldBeEmpty)
		})

		Convey("returns RateLimited when the bucket rejects", func() {
			limiter := &fakeMutationRateLimiter{rejectAt: 1}
			h := newHandler(limiter, newEnabledScope())

			err := h.checkMutationRateLimit(ctx, 1)

			So(err, ShouldNotBeNil)
			apiErr := apierrors.AsAPIErrorWithContext(ctx, err)
			So(apiErr, ShouldNotBeNil)
			So(apiErr.Kind.Name, ShouldEqual, apierrors.TooManyRequest)
			So(apiErr.Kind.Reason, ShouldEqual, "RateLimited")
		})

		Convey("is a no-op when the scope config is absent", func() {
			// Only reachable before SetFieldDefaults has run; at runtime the
			// chain is always populated.
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, nil)

			So(h.checkMutationRateLimit(ctx, 3), ShouldBeNil)
			So(limiter.calls, ShouldBeEmpty)
		})
	})
}
