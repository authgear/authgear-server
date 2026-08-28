package handler

import (
	"context"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
)

type stubIDTokenIssuer struct{}

func (stubIDTokenIssuer) Iss() string { return "https://app.authgear.example.com" }
func (stubIDTokenIssuer) PrepareIDToken(ctx context.Context, opts oidc.PrepareIDTokenOptions) (*oidc.PrepareIDTokenResult, error) {
	panic("not implemented")
}
func (stubIDTokenIssuer) MakeIDTokenFromPreparationResult(ctx context.Context, opts oidc.MakeIDTokenFromPreparationResultOptions) (string, error) {
	panic("not implemented")
}
func (stubIDTokenIssuer) VerifyIDToken(idToken string) (jwt.Token, error) {
	panic("not implemented")
}

type stubResourceScopeService struct {
	resource *resourcescope.Resource
	scopes   []*resourcescope.Scope
	err      error
}

func (s *stubResourceScopeService) GetResourceByURIForThirdPartyAccess(ctx context.Context, uri string) (*resourcescope.Resource, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.resource, nil
}

func (s *stubResourceScopeService) ListScopesForThirdPartyAccess(ctx context.Context, resourceID string) ([]*resourcescope.Scope, error) {
	return s.scopes, nil
}

func TestAuthorizationHandlerValidateResource(t *testing.T) {
	Convey("AuthorizationHandler.validateResource", t, func() {
		dynamicThirdPartyClient := &config.OAuthClientConfig{
			ClientID:        "dynamic-third-party-client",
			ApplicationType: config.OAuthClientApplicationTypeDynamicThirdParty,
			IsDynamic:       true,
		}
		staticThirdPartyClient := &config.OAuthClientConfig{
			ClientID:        "static-third-party-client",
			ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp,
		}
		dynamicFirstPartyClient := &config.OAuthClientConfig{
			ClientID:        "dynamic-first-party-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
			IsDynamic:       true,
		}
		spaClient := &config.OAuthClientConfig{
			ClientID:        "spa-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}

		Convey("no resource requested returns nil, nil", func() {
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			scopes, err := h.validateResource(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{})
			So(err, ShouldBeNil)
			So(scopes, ShouldBeNil)
		})

		Convey("resource prefixed by the project endpoint is invalid_target", func() {
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			scopes, err := h.validateResource(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://app.authgear.example.com/oauth2/userinfo",
			})
			So(scopes, ShouldBeNil)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "resource URI must not be a prefixed by authgear endpoint"))
		})

		Convey("dynamic third-party client with a policy-enabled resource returns its allowed scopes", func() {
			h := &AuthorizationHandler{
				IDTokenIssuer: stubIDTokenIssuer{},
				Database:      &db.MockHandle{},
				ResourceScopeService: &stubResourceScopeService{
					resource: &resourcescope.Resource{ID: "resource-id", ResourceURI: "https://api.example.com/orders"},
					scopes: []*resourcescope.Scope{
						{Scope: "read:orders"},
						{Scope: "write:orders"},
					},
				},
			}
			scopes, err := h.validateResource(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(err, ShouldBeNil)
			So(scopes, ShouldResemble, []string{"read:orders", "write:orders"})
		})

		Convey("dynamic third-party client with a policy-disabled (not found) resource is invalid_target", func() {
			h := &AuthorizationHandler{
				IDTokenIssuer: stubIDTokenIssuer{},
				Database:      &db.MockHandle{},
				ResourceScopeService: &stubResourceScopeService{
					err: resourcescope.ErrResourceNotFound,
				},
			}
			scopes, err := h.validateResource(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/secret",
			})
			So(scopes, ShouldBeNil)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "resource not found or not accessible to third-party clients"))
		})

		Convey("static third-party client requesting a resource is invalid_target, not routed through the dynamic access policy", func() {
			// Regression test: IsThirdParty() alone is also true for a static
			// third_party_app client, but the allow_dynamic_third_party_client_access
			// policy must only ever be reachable by a dynamically-resolved
			// (DCR/CIMD) client. A static third-party client requesting a
			// resource must fall through to the same "not permitted" error
			// as spa/native/confidential below, not be granted access via
			// that policy.
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			scopes, err := h.validateResource(context.Background(), staticThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(scopes, ShouldBeNil)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "this client is not permitted to use the resource parameter"))
		})

		Convey("dynamic first-party client requesting a resource is invalid_target, not routed through the dynamic access policy", func() {
			// Regression test for the other half of the fix: IsDynamicClient()
			// alone is also true for a dynamic first-party client (Kind ==
			// FIRST_PARTY resolves to the ordinary "spa"/"native"
			// ApplicationType, not the synthetic DynamicThirdParty one), but
			// first-party support for the resource parameter is deferred,
			// same as for static first-party clients -- both IsDynamicClient
			// and IsThirdParty must hold for the dynamic-access-policy path.
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			scopes, err := h.validateResource(context.Background(), dynamicFirstPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(scopes, ShouldBeNil)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "this client is not permitted to use the resource parameter"))
		})

		Convey("static spa/native/confidential client requesting any resource is invalid_target", func() {
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			scopes, err := h.validateResource(context.Background(), spaClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(scopes, ShouldBeNil)
			So(err, ShouldResemble, protocol.NewError("invalid_target", "this client is not permitted to use the resource parameter"))
		})
	})
}

func TestAuthorizationHandlerResourceScopeDisplayNames(t *testing.T) {
	Convey("AuthorizationHandler.resourceScopeDisplayNames", t, func() {
		dynamicThirdPartyClient := &config.OAuthClientConfig{
			ClientID:        "dynamic-third-party-client",
			ApplicationType: config.OAuthClientApplicationTypeDynamicThirdParty,
			IsDynamic:       true,
		}
		spaClient := &config.OAuthClientConfig{
			ClientID:        "spa-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		desc := "Read your orders"

		Convey("no resource requested returns nil", func() {
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			names := h.resourceScopeDisplayNames(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{})
			So(names, ShouldBeNil)
		})

		Convey("a client type not eligible for the resource parameter returns nil", func() {
			h := &AuthorizationHandler{IDTokenIssuer: stubIDTokenIssuer{}}
			names := h.resourceScopeDisplayNames(context.Background(), spaClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(names, ShouldBeNil)
		})

		Convey("falls back to the raw scope name when Description is unset", func() {
			h := &AuthorizationHandler{
				IDTokenIssuer: stubIDTokenIssuer{},
				Database:      &db.MockHandle{},
				ResourceScopeService: &stubResourceScopeService{
					resource: &resourcescope.Resource{ID: "resource-id", ResourceURI: "https://api.example.com/orders"},
					scopes: []*resourcescope.Scope{
						{Scope: "read:orders", Description: &desc},
						{Scope: "write:orders"},
					},
				},
			}
			names := h.resourceScopeDisplayNames(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/orders",
			})
			So(names, ShouldResemble, map[string]string{
				"read:orders":  "Read your orders",
				"write:orders": "write:orders",
			})
		})

		Convey("a lookup failure returns nil rather than an error", func() {
			h := &AuthorizationHandler{
				IDTokenIssuer: stubIDTokenIssuer{},
				Database:      &db.MockHandle{},
				ResourceScopeService: &stubResourceScopeService{
					err: resourcescope.ErrResourceNotFound,
				},
			}
			names := h.resourceScopeDisplayNames(context.Background(), dynamicThirdPartyClient, protocol.AuthorizationRequest{
				"resource": "https://api.example.com/secret",
			})
			So(names, ShouldBeNil)
		})
	})
}
