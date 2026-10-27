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

func TestNewGrantScopes(t *testing.T) {
	Convey("newGrantScopes", t, func() {
		ctx := context.Background()
		issuer := stubIDTokenIssuer{}.Iss()
		spaClient := &config.OAuthClientConfig{
			ClientID:        "spa-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		defaults := []string{"openid", "offline_access", "https://authgear.com/scopes/full-access"}

		svc := &stubResourceScopeService{
			resource: &resourcescope.Resource{
				ID:           "resource-id",
				ResourceURI:  "https://api.example.com/orders",
				AccessPolicy: model.AccessPolicy{AllowStaticFirstPartyClientAccess: true},
			},
			scopes: []*resourcescope.Scope{
				{Scope: "read:orders", AccessPolicy: model.AccessPolicy{AllowStaticFirstPartyClientAccess: true}},
				// Known, but not opened to static first-party clients.
				{Scope: "delete:orders"},
			},
		}

		Convey("no resource and no scope returns the defaults", func() {
			resourceID, scopes, err := newGrantScopes(ctx, svc, issuer, spaClient, "", nil, defaults)
			So(err, ShouldBeNil)
			So(resourceID, ShouldEqual, "")
			So(scopes, ShouldResemble, defaults)
		})

		Convey("a resource scope without resource is invalid_scope", func() {
			_, _, err := newGrantScopes(ctx, svc, issuer, spaClient, "", []string{"openid", "read:orders"}, defaults)
			So(err, ShouldNotBeNil)
			So(err.(*protocol.OAuthProtocolError).Type(), ShouldEqual, "invalid_scope")
		})

		Convey("a resource under the project endpoint is invalid_target", func() {
			_, _, err := newGrantScopes(ctx, svc, issuer, spaClient, issuer+"/oauth2/userinfo", nil, defaults)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "resource URI must not be a prefixed by authgear endpoint"))
		})

		Convey("a resource that does not allow the client is invalid_target", func() {
			denied := &stubResourceScopeService{
				resource: &resourcescope.Resource{ID: "resource-id", ResourceURI: "https://api.example.com/orders"},
			}
			_, _, err := newGrantScopes(ctx, denied, issuer, spaClient, "https://api.example.com/orders", nil, defaults)
			So(err, ShouldNotBeNil)
			So(err.(*protocol.OAuthProtocolError).Type(), ShouldEqual, "invalid_target")
		})

		Convey("a resource with scopes returns the id and drops scopes the client is not granted", func() {
			resourceID, scopes, err := newGrantScopes(ctx, svc, issuer, spaClient, "https://api.example.com/orders",
				[]string{"openid", "offline_access", "https://authgear.com/scopes/full-access", "read:orders", "delete:orders"}, defaults)
			So(err, ShouldBeNil)
			So(resourceID, ShouldEqual, "resource-id")
			So(scopes, ShouldResemble, []string{"openid", "offline_access", "https://authgear.com/scopes/full-access", "read:orders"})
		})

		Convey("an unknown scope is invalid_scope", func() {
			_, _, err := newGrantScopes(ctx, svc, issuer, spaClient, "https://api.example.com/orders",
				[]string{"openid", "unknown:scope"}, defaults)
			So(err, ShouldNotBeNil)
			So(err.(*protocol.OAuthProtocolError).Type(), ShouldEqual, "invalid_scope")
		})
	})
}

func TestCheckOfflineFullAccessScopes(t *testing.T) {
	Convey("checkOfflineFullAccessScopes", t, func() {
		So(checkOfflineFullAccessScopes([]string{"openid", "offline_access", "https://authgear.com/scopes/full-access"}), ShouldBeNil)

		err := checkOfflineFullAccessScopes([]string{"openid", "https://authgear.com/scopes/full-access"})
		So(err, ShouldResemble, protocol.NewError("invalid_scope", "offline_access and full-access must be requested"))

		err = checkOfflineFullAccessScopes([]string{"openid", "offline_access"})
		So(err, ShouldResemble, protocol.NewError("invalid_scope", "offline_access and full-access must be requested"))
	})
}

func TestDerivedGrantResourceScopes(t *testing.T) {
	Convey("derivedGrantResourceScopes", t, func() {
		ctx := context.Background()
		const orders = "https://api.example.com/orders"

		original := &config.OAuthClientConfig{
			ClientID:        "static-first-party",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		thirdParty := &config.OAuthClientConfig{
			ClientID:        "static-third-party",
			ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp,
		}
		// The original refresh token holds project-level scopes and three
		// scopes of the resource.
		originals := map[string][]string{
			orders: {"openid", "offline_access", "read:orders", "write:orders", "delete:orders"},
		}

		svc := func(resource model.AccessPolicy) *stubResourceScopeService {
			return &stubResourceScopeService{
				resource: &resourcescope.Resource{ID: "resource-id", ResourceURI: orders, AccessPolicy: resource},
				scopes: []*resourcescope.Scope{
					{Scope: "read:orders", AccessPolicy: model.AccessPolicy{AllowStaticFirstPartyClientAccess: true, AllowStaticThirdPartyClientAccess: true}},
					{Scope: "write:orders", AccessPolicy: model.AccessPolicy{AllowStaticFirstPartyClientAccess: true}},
					{Scope: "delete:orders", AccessPolicy: model.AccessPolicy{AllowStaticThirdPartyClientAccess: true}},
				},
			}
		}
		both := model.AccessPolicy{AllowStaticFirstPartyClientAccess: true, AllowStaticThirdPartyClientAccess: true}

		Convey("no resource carries nothing", func() {
			resourceID, scopes, err := derivedGrantResourceScopes(ctx, svc(both), original, thirdParty, originals, "")
			So(err, ShouldBeNil)
			So(resourceID, ShouldEqual, "")
			So(scopes, ShouldBeNil)
		})

		Convey("a resource the original refresh token is not bound to is invalid_target", func() {
			_, _, err := derivedGrantResourceScopes(ctx, svc(both), original, thirdParty, map[string][]string{"": {"openid"}}, orders)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "resource is not bound to the original session"))
		})

		Convey("only scopes both clients are granted are carried, and project scopes are not", func() {
			resourceID, scopes, err := derivedGrantResourceScopes(ctx, svc(both), original, thirdParty, originals, orders)
			So(err, ShouldBeNil)
			So(resourceID, ShouldEqual, "resource-id")
			// write:orders is not granted to the third-party client, and
			// delete:orders is not granted to the original client.
			So(scopes, ShouldResemble, []string{"read:orders"})
		})

		Convey("a resource that does not allow the new client is invalid_target", func() {
			_, _, err := derivedGrantResourceScopes(ctx, svc(model.AccessPolicy{AllowStaticFirstPartyClientAccess: true}), original, thirdParty, originals, orders)
			So(err, ShouldNotBeNil)
			So(err.(*protocol.OAuthProtocolError).Type(), ShouldEqual, "invalid_target")
		})

		Convey("a resource that does not allow the original client is invalid_target", func() {
			_, _, err := derivedGrantResourceScopes(ctx, svc(model.AccessPolicy{AllowStaticThirdPartyClientAccess: true}), original, thirdParty, originals, orders)
			So(err, ShouldNotBeNil)
			So(err.(*protocol.OAuthProtocolError).Type(), ShouldEqual, "invalid_target")
		})
	})
}
