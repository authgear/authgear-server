package transport

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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
		PerProject: &config.RateLimitConfig{Enabled: &enabled, Period: "1m", Burst: 300},
		PerIP:      &config.RateLimitConfig{Enabled: &enabled, Period: "1m", Burst: 150},
	}
}

func newHandler(limiter *fakeMutationRateLimiter, scope *config.AdminAPIRateLimitsMutationScopeFeatureConfig) *GraphQLHandler {
	return &GraphQLHandler{
		RateLimiter: limiter,
		RemoteIP:    httputil.RemoteIP("1.2.3.4"),
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

		Convey("charges one token per mutation field to both buckets", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 3), ShouldBeNil)

			So(limiter.ns, ShouldResemble, []int{3, 3})
			So(limiter.calls, ShouldHaveLength, 2)
		})

		Convey("charges per-IP before per-project", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 1), ShouldBeNil)

			So(limiter.calls[0].RateLimitName, ShouldEqual, ratelimit.RateLimitAdminAPIMutationAllPerIP)
			So(limiter.calls[1].RateLimitName, ShouldEqual, ratelimit.RateLimitAdminAPIMutationAllPerProject)
		})

		Convey("keys the per-IP bucket on the remote IP", func() {
			limiter := &fakeMutationRateLimiter{}
			h := newHandler(limiter, newEnabledScope())

			So(h.checkMutationRateLimit(ctx, 1), ShouldBeNil)

			So(limiter.calls[0].Arguments, ShouldResemble, []string{"1.2.3.4"})
			So(limiter.calls[1].Arguments, ShouldBeEmpty)
		})

		Convey("returns RateLimited and stops when the per-IP bucket rejects", func() {
			limiter := &fakeMutationRateLimiter{rejectAt: 1}
			h := newHandler(limiter, newEnabledScope())

			err := h.checkMutationRateLimit(ctx, 1)

			So(err, ShouldNotBeNil)
			apiErr := apierrors.AsAPIErrorWithContext(ctx, err)
			So(apiErr, ShouldNotBeNil)
			So(apiErr.Kind.Name, ShouldEqual, apierrors.TooManyRequest)
			So(apiErr.Kind.Reason, ShouldEqual, "RateLimited")
			// The per-project bucket is never charged for a request the
			// per-IP bucket already rejected.
			So(limiter.calls, ShouldHaveLength, 1)
		})

		Convey("returns RateLimited when the per-project bucket rejects", func() {
			limiter := &fakeMutationRateLimiter{rejectAt: 2}
			h := newHandler(limiter, newEnabledScope())

			err := h.checkMutationRateLimit(ctx, 1)

			So(err, ShouldNotBeNil)
			So(limiter.calls, ShouldHaveLength, 2)
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
