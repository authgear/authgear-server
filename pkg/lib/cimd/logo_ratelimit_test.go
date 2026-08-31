package cimd

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func TestNewBucketSpecCIMDLogoPerClient(t *testing.T) {
	Convey("NewBucketSpecCIMDLogoPerClient", t, func() {
		clientID := "https://mcp-client.example.com/oauth/client-metadata.json"
		spec := NewBucketSpecCIMDLogoPerClient(clientID)

		So(spec.Name, ShouldEqual, ratelimit.OAuthCIMDLogoPerClient)
		So(spec.RateLimitName, ShouldEqual, ratelimit.RateLimitOAuthCIMDLogoPerClient)
		So(spec.RateLimitGroup, ShouldEqual, ratelimit.RateLimitGroupOAuthCIMDLogo)
		So(spec.Enabled, ShouldBeTrue)
		So(spec.Period, ShouldEqual, time.Minute)
		So(spec.Burst, ShouldEqual, 2)

		Convey("the client_id argument is hashed, not passed raw -- a raw URL contains ':' and BucketSpec.Key() joins arguments with ':'", func() {
			So(spec.Arguments, ShouldResemble, []string{crypto.SHA256String(clientID)})
			So(spec.Arguments, ShouldNotContain, clientID)
		})

		Convey("different client_ids hash to different arguments", func() {
			other := NewBucketSpecCIMDLogoPerClient("https://other-client.example.com/metadata.json")
			So(other.Arguments, ShouldNotResemble, spec.Arguments)
		})
	})
}
