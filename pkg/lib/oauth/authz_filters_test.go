package oauth_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type stubClientResolver struct {
	clients map[string]*config.OAuthClientConfig
	gotCtx  []context.Context
}

func (s *stubClientResolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	s.gotCtx = append(s.gotCtx, ctx)
	return s.clients[clientID]
}

func TestKeepThirdPartyAuthorizationFilter(t *testing.T) {
	Convey("KeepThirdPartyAuthorizationFilter.Keep", t, func() {
		resolver := &stubClientResolver{
			clients: map[string]*config.OAuthClientConfig{
				"static-third-party": {
					ClientID:        "static-third-party",
					ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp,
				},
				"static-spa": {
					ClientID:        "static-spa",
					ApplicationType: config.OAuthClientApplicationTypeSPA,
				},
				"static-native": {
					ClientID:        "static-native",
					ApplicationType: config.OAuthClientApplicationTypeNative,
				},
				"static-traditional-webapp": {
					ClientID:        "static-traditional-webapp",
					ApplicationType: config.OAuthClientApplicationTypeTraditionalWeb,
				},
				"static-confidential": {
					ClientID:        "static-confidential",
					ApplicationType: config.OAuthClientApplicationTypeConfidential,
				},
				"static-m2m": {
					ClientID:        "static-m2m",
					ApplicationType: config.OAuthClientApplicationTypeM2M,
				},
				"dynamic-third-party": {
					ClientID:        "dynamic-third-party",
					ApplicationType: config.OAuthClientApplicationTypeDynamicThirdParty,
					DynamicSource:   "CIMD",
				},
				"dynamic-first-party": {
					ClientID:        "dynamic-first-party",
					ApplicationType: config.OAuthClientApplicationTypeSPA,
					DynamicSource:   "CIMD",
				},
			},
		}
		filter := oauth.NewKeepThirdPartyAuthorizationFilter(resolver)
		ctx := context.Background()

		cases := []struct {
			clientID string
			keep     bool
		}{
			{"static-third-party", true},
			{"static-spa", false},
			{"static-native", false},
			{"static-traditional-webapp", false},
			{"static-confidential", false},
			{"static-m2m", false},
			// The regression this fixes: a DCR/CIMD-resolved third-party
			// client used to be dropped because it was never in the
			// static-client-id set built from authgear.yaml.
			{"dynamic-third-party", true},
			{"dynamic-first-party", false},
			{"unknown-client-id", false},
		}

		for _, tc := range cases {
			Convey(tc.clientID, func() {
				authz := &oauth.Authorization{ClientID: tc.clientID}
				So(filter.Keep(ctx, authz), ShouldEqual, tc.keep)
			})
		}

		Convey("ctx is threaded through to ResolveClient", func() {
			type ctxKey struct{}
			ctx := context.WithValue(context.Background(), ctxKey{}, "marker")
			authz := &oauth.Authorization{ClientID: "static-third-party"}
			filter.Keep(ctx, authz)
			So(len(resolver.gotCtx), ShouldBeGreaterThan, 0)
			last := resolver.gotCtx[len(resolver.gotCtx)-1]
			So(last.Value(ctxKey{}), ShouldEqual, "marker")
		})
	})
}
