package handler

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
)

func TestResourceScopesForIssuance(t *testing.T) {
	Convey("resourceScopesForIssuance", t, func() {
		ctx := context.Background()
		spaClient := &config.OAuthClientConfig{
			ClientID:        "spa-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		granted := []string{"openid", "offline_access", "read:orders", "write:orders"}

		svc := func(resourceAllowed bool, allowedScopes ...string) *stubResourceScopeService {
			policy := func(allowed bool) model.AccessPolicy {
				return model.AccessPolicy{AllowStaticFirstPartyClientAccess: allowed}
			}
			scopes := []*resourcescope.Scope{
				{Scope: "read:orders", AccessPolicy: policy(false)},
				{Scope: "write:orders", AccessPolicy: policy(false)},
			}
			for _, s := range scopes {
				for _, a := range allowedScopes {
					if s.Scope == a {
						s.AccessPolicy = policy(true)
					}
				}
			}
			return &stubResourceScopeService{
				resource: &resourcescope.Resource{
					ID:           "resource-id",
					ResourceURI:  "https://api.example.com/orders",
					AccessPolicy: policy(resourceAllowed),
				},
				scopes: scopes,
			}
		}

		Convey("unbound token drops every resource scope and keeps the rest", func() {
			issued, err := resourceScopesForIssuance(ctx, svc(true, "read:orders"), spaClient, "", granted)
			So(err, ShouldBeNil)
			So(issued, ShouldResemble, []string{"openid", "offline_access"})
		})

		Convey("bound token drops a scope the policy no longer allows", func() {
			issued, err := resourceScopesForIssuance(ctx, svc(true, "read:orders"), spaClient, "https://api.example.com/orders", granted)
			So(err, ShouldBeNil)
			So(issued, ShouldResemble, []string{"openid", "offline_access", "read:orders"})
		})

		Convey("bound token fails when the resource no longer allows the client", func() {
			_, err := resourceScopesForIssuance(ctx, svc(false, "read:orders"), spaClient, "https://api.example.com/orders", granted)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "resource not found or not accessible to this client"))
		})
	})
}
