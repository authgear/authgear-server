package cimd

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// stubLimiter records every spec it is asked to Allow, in order, and returns
// canned results keyed by BucketName.
type stubLimiter struct {
	specs   []ratelimit.BucketSpec
	failFor map[ratelimit.BucketName]*ratelimit.FailedReservation
	errFor  map[ratelimit.BucketName]error
}

func (s *stubLimiter) Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error) {
	s.specs = append(s.specs, spec)
	if err, ok := s.errFor[spec.Name]; ok {
		return nil, err
	}
	if failed, ok := s.failFor[spec.Name]; ok {
		return failed, nil
	}
	return nil, nil
}

func testFetchRateLimitsConfig(perProjectEnabled bool, perProjectPeriod string, perProjectBurst int, perIPEnabled bool, perIPPeriod string, perIPBurst int) *config.OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig {
	return &config.OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig{
		PerProject: &config.RateLimitConfig{
			Enabled: &perProjectEnabled,
			Period:  config.DurationString(perProjectPeriod),
			Burst:   perProjectBurst,
		},
		PerIP: &config.RateLimitConfig{
			Enabled: &perIPEnabled,
			Period:  config.DurationString(perIPPeriod),
			Burst:   perIPBurst,
		},
	}
}

func testRateLimitsConfig(perProjectEnabled bool, perProjectPeriod string, perProjectBurst int, perIPEnabled bool, perIPPeriod string, perIPBurst int) *config.OAuthClientIDMetadataDocumentRateLimitsFeatureConfig {
	return &config.OAuthClientIDMetadataDocumentRateLimitsFeatureConfig{
		Fetch: testFetchRateLimitsConfig(perProjectEnabled, perProjectPeriod, perProjectBurst, perIPEnabled, perIPPeriod, perIPBurst),
	}
}

func TestRateLimiterCheckFetchAllowed(t *testing.T) {
	Convey("RateLimiter.CheckFetchAllowed", t, func() {
		Convey("requests exactly two specs, per-IP then per-project, with values from the supplied feature config", func() {
			rateLimits := testRateLimitsConfig(true, "1m", 10, true, "30s", 5)
			limiter := &stubLimiter{}
			r := &RateLimiter{
				Limiter:  limiter,
				RemoteIP: "203.0.113.9",
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						RateLimits: rateLimits,
					},
				},
			}

			err := r.CheckFetchAllowed(context.Background())
			So(err, ShouldBeNil)
			So(len(limiter.specs), ShouldEqual, 2)

			perIP := limiter.specs[0]
			So(perIP.Name, ShouldEqual, ratelimit.OAuthClientIDMetadataDocumentFetchPerIP)
			So(perIP.RateLimitName, ShouldEqual, ratelimit.RateLimitOAuthClientIDMetadataDocumentFetchPerIP)
			So(perIP.RateLimitGroup, ShouldEqual, ratelimit.RateLimitGroupOAuthClientIDMetadataDocumentFetch)
			So(perIP.Enabled, ShouldBeTrue)
			So(perIP.Period, ShouldEqual, 30*time.Second)
			So(perIP.Burst, ShouldEqual, 5)
			So(perIP.Arguments, ShouldResemble, []string{"203.0.113.9"})

			perProject := limiter.specs[1]
			So(perProject.Name, ShouldEqual, ratelimit.OAuthClientIDMetadataDocumentFetchPerProject)
			So(perProject.RateLimitName, ShouldEqual, ratelimit.RateLimitOAuthClientIDMetadataDocumentFetchPerProject)
			So(perProject.RateLimitGroup, ShouldEqual, ratelimit.RateLimitGroupOAuthClientIDMetadataDocumentFetch)
			So(perProject.Enabled, ShouldBeTrue)
			So(perProject.Period, ShouldEqual, time.Minute)
			So(perProject.Burst, ShouldEqual, 10)
			So(perProject.Arguments, ShouldBeEmpty)
		})

		Convey("a failed per-IP reservation short-circuits before the per-project bucket is ever checked", func() {
			fetch := testFetchRateLimitsConfig(true, "1m", 10, true, "1m", 5)
			rateLimits := &config.OAuthClientIDMetadataDocumentRateLimitsFeatureConfig{Fetch: fetch}
			spec := NewBucketSpecCIMDFetchPerIP(fetch, "203.0.113.9")
			limiter := &stubLimiter{
				failFor: map[ratelimit.BucketName]*ratelimit.FailedReservation{
					ratelimit.OAuthClientIDMetadataDocumentFetchPerIP: ratelimit.NewFailedReservation(spec),
				},
			}
			r := &RateLimiter{
				Limiter:  limiter,
				RemoteIP: "203.0.113.9",
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						RateLimits: rateLimits,
					},
				},
			}

			err := r.CheckFetchAllowed(context.Background())
			So(err, ShouldNotBeNil)
			So(apierrors.IsKind(err, ratelimit.RateLimited), ShouldBeTrue)
			So(len(limiter.specs), ShouldEqual, 1)
			So(limiter.specs[0].Name, ShouldEqual, ratelimit.OAuthClientIDMetadataDocumentFetchPerIP)
		})

		Convey("a disabled per-IP bucket is still passed to Allow, with Enabled: false, and the limiter itself is trusted to short-circuit", func() {
			rateLimits := testRateLimitsConfig(true, "1m", 10, false, "1m", 5)
			limiter := &stubLimiter{}
			r := &RateLimiter{
				Limiter:  limiter,
				RemoteIP: "203.0.113.9",
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						RateLimits: rateLimits,
					},
				},
			}

			err := r.CheckFetchAllowed(context.Background())
			So(err, ShouldBeNil)
			So(len(limiter.specs), ShouldEqual, 2)
			So(limiter.specs[0].Enabled, ShouldBeFalse)
			So(limiter.specs[1].Enabled, ShouldBeTrue)
		})

		Convey("a transport error from Allow is returned unchanged", func() {
			rateLimits := testRateLimitsConfig(true, "1m", 10, true, "1m", 5)
			wantErr := errors.New("redis: connection refused")
			limiter := &stubLimiter{
				errFor: map[ratelimit.BucketName]error{
					ratelimit.OAuthClientIDMetadataDocumentFetchPerIP: wantErr,
				},
			}
			r := &RateLimiter{
				Limiter:  limiter,
				RemoteIP: "203.0.113.9",
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						RateLimits: rateLimits,
					},
				},
			}

			err := r.CheckFetchAllowed(context.Background())
			So(errors.Is(err, wantErr), ShouldBeTrue)
			So(len(limiter.specs), ShouldEqual, 1)
		})
	})
}
