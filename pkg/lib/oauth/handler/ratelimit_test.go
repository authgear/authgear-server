package handler

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func TestNewBucketSpecOAuthRegister(t *testing.T) {
	Convey("NewBucketSpecOAuthRegisterPerIP/PerProject", t, func() {
		Convey("with no project configuration, falls back to the built-in defaults", func() {
			rateLimits := &config.OAuthDynamicClientRegistrationRateLimitsConfig{}
			// SetFieldDefaults, not SetDefaults() directly: SetDefaults() assumes
			// PerIP/PerProject are already allocated, which is only true once the
			// generic walker has recursed into them first -- exactly like a real
			// config.Parse() would, and exactly why this is not "rateLimits.SetDefaults()".
			config.SetFieldDefaults(rateLimits)

			spec := NewBucketSpecOAuthRegisterPerIP(rateLimits, "127.0.0.1")
			So(spec.Enabled, ShouldBeTrue)
			So(spec.Period, ShouldEqual, time.Minute)
			So(spec.Burst, ShouldEqual, 10)
			So(spec.RateLimitName, ShouldEqual, ratelimit.RateLimitOAuthRegisterPerIP)
			So(spec.RateLimitGroup, ShouldEqual, ratelimit.RateLimitGroupOAuthRegister)
			So(spec.Name, ShouldEqual, ratelimit.OAuthRegisterPerIP)
			So(spec.Arguments, ShouldResemble, []string{"127.0.0.1"})

			projectSpec := NewBucketSpecOAuthRegisterPerProject(rateLimits)
			So(projectSpec.Enabled, ShouldBeTrue)
			So(projectSpec.Period, ShouldEqual, time.Hour)
			So(projectSpec.Burst, ShouldEqual, 1000)
			So(projectSpec.RateLimitName, ShouldEqual, ratelimit.RateLimitOAuthRegisterPerProject)
			So(projectSpec.RateLimitGroup, ShouldEqual, ratelimit.RateLimitGroupOAuthRegister)
			So(projectSpec.Name, ShouldEqual, ratelimit.OAuthRegisterPerProject)
		})

		Convey("with a project-configured override, uses the configured rate", func() {
			burst := 5
			rateLimits := &config.OAuthDynamicClientRegistrationRateLimitsConfig{
				PerIP: &config.RateLimitConfig{Enabled: new(true), Period: "30s", Burst: burst},
			}
			config.SetFieldDefaults(rateLimits)

			spec := NewBucketSpecOAuthRegisterPerIP(rateLimits, "127.0.0.1")
			So(spec.Period, ShouldEqual, 30*time.Second)
			So(spec.Burst, ShouldEqual, burst)

			// PerProject was left unconfigured -- still falls back to the default.
			projectSpec := NewBucketSpecOAuthRegisterPerProject(rateLimits)
			So(projectSpec.Period, ShouldEqual, time.Hour)
			So(projectSpec.Burst, ShouldEqual, 1000)
		})
	})
}
