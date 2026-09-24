package graphql

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func TestBuildAuthorizationScopes(t *testing.T) {
	Convey("buildAuthorizationScopes", t, func() {
		billing := &model.Resource{ResourceURI: "https://api.example.com/billing"}
		billing.ID = "resource-billing"
		inventory := &model.Resource{ResourceURI: "https://api.example.com/inventory"}
		inventory.ID = "resource-inventory"

		readOrders := "Read orders and their line items"
		resources := map[string]*model.Resource{
			billing.ID:   billing,
			inventory.ID: inventory,
		}

		Convey("a project-level scope carries no resource", func() {
			result := buildAuthorizationScopes(
				[]string{"openid", "profile"},
				map[string][]*model.Scope{},
				resources,
			)

			So(result, ShouldHaveLength, 2)
			for _, r := range result {
				So(r.Kind, ShouldEqual, authorizationScopeKindProject)
				So(r.Resource, ShouldBeNil)
				So(r.Description, ShouldBeNil)
			}
		})

		Convey("a resource scope carries its resource and description", func() {
			result := buildAuthorizationScopes(
				[]string{"read:orders"},
				map[string][]*model.Scope{
					"read:orders": {
						{ResourceID: billing.ID, Scope: "read:orders", Description: &readOrders},
					},
				},
				resources,
			)

			So(result, ShouldHaveLength, 1)
			So(result[0].Kind, ShouldEqual, authorizationScopeKindResource)
			So(result[0].Resource, ShouldEqual, billing)
			So(*result[0].Description, ShouldEqual, readOrders)
		})

		// The grant stores names only, so a name two resources define is
		// genuinely ambiguous. Reporting both is what makes that visible.
		Convey("a name defined by two resources yields one entry per resource, ordered by URI", func() {
			result := buildAuthorizationScopes(
				[]string{"read:orders"},
				map[string][]*model.Scope{
					// Deliberately not in URI order.
					"read:orders": {
						{ResourceID: inventory.ID, Scope: "read:orders"},
						{ResourceID: billing.ID, Scope: "read:orders"},
					},
				},
				resources,
			)

			So(result, ShouldHaveLength, 2)
			So(result[0].Resource, ShouldEqual, billing)
			So(result[1].Resource, ShouldEqual, inventory)
		})

		// Filing it under PROJECT would claim it is a project-level scope.
		Convey("a scope the project no longer defines stays RESOURCE with no resource", func() {
			result := buildAuthorizationScopes(
				[]string{"read:legacy-thing"},
				map[string][]*model.Scope{},
				resources,
			)

			So(result, ShouldHaveLength, 1)
			So(result[0].Kind, ShouldEqual, authorizationScopeKindResource)
			So(result[0].Resource, ShouldBeNil)
		})

		Convey("the granted order is preserved", func() {
			result := buildAuthorizationScopes(
				[]string{"openid", "read:orders", "profile"},
				map[string][]*model.Scope{
					"read:orders": {{ResourceID: billing.ID, Scope: "read:orders"}},
				},
				resources,
			)

			So(result, ShouldHaveLength, 3)
			So(result[0].Scope, ShouldEqual, "openid")
			So(result[1].Scope, ShouldEqual, "read:orders")
			So(result[2].Scope, ShouldEqual, "profile")
		})
	})
}

type stubResourceLoader struct {
	resources map[string]*model.Resource
	loaded    []string
}

func (l *stubResourceLoader) Load(ctx context.Context, key any) *graphqlutil.Lazy {
	id := key.(string)
	l.loaded = append(l.loaded, id)
	return graphqlutil.NewLazy(func() (any, error) {
		return l.resources[id], nil
	})
}

func (l *stubResourceLoader) LoadMany(ctx context.Context, keys []any) *graphqlutil.Lazy {
	panic("not used")
}
func (l *stubResourceLoader) Clear(key any)            {}
func (l *stubResourceLoader) ClearAll()                {}
func (l *stubResourceLoader) Prime(key any, value any) {}

type stubResourceScopeFacade struct {
	ResourceScopeFacade
	scopes        []*model.Scope
	askedForNames []string
}

func (f *stubResourceScopeFacade) ListScopesByNames(ctx context.Context, names []string) ([]*model.Scope, error) {
	f.askedForNames = names
	return f.scopes, nil
}

func TestResolveScopes(t *testing.T) {
	Convey("resolveScopes", t, func() {
		ctx := context.Background()
		mcp := &model.Resource{ResourceURI: "https://mcp.example.com", Name: new("MCP Server")}
		mcp.ID = "resource-mcp"
		desc := "Run tools exposed by the server"

		loader := &stubResourceLoader{resources: map[string]*model.Resource{mcp.ID: mcp}}
		facade := &stubResourceScopeFacade{scopes: []*model.Scope{
			{ResourceID: mcp.ID, Scope: "execute:tools", Description: &desc},
		}}
		gqlCtx := &Context{Resources: loader, ResourceScopeFacade: facade}

		Convey("only resource scope names are looked up", func() {
			_, err := resolveScopes(ctx, gqlCtx, []string{"openid", "profile", "execute:tools"})
			So(err, ShouldBeNil)
			So(facade.askedForNames, ShouldResemble, []string{"execute:tools"})
		})

		Convey("a resource is loaded once even when two scopes share it", func() {
			facade.scopes = []*model.Scope{
				{ResourceID: mcp.ID, Scope: "execute:tools"},
				{ResourceID: mcp.ID, Scope: "read:resources"},
			}

			result, err := resolveScopes(ctx, gqlCtx, []string{"execute:tools", "read:resources"})
			So(err, ShouldBeNil)
			So(result, ShouldHaveLength, 2)
			So(loader.loaded, ShouldResemble, []string{mcp.ID})
		})

		Convey("a grant with no resource scope makes no lookup at all", func() {
			result, err := resolveScopes(ctx, gqlCtx, []string{"openid", "offline_access"})
			So(err, ShouldBeNil)
			So(result, ShouldHaveLength, 2)
			So(facade.askedForNames, ShouldBeNil)
			So(loader.loaded, ShouldBeNil)
		})

		Convey("the resolved entry carries the resource and description", func() {
			result, err := resolveScopes(ctx, gqlCtx, []string{"execute:tools"})
			So(err, ShouldBeNil)
			So(result, ShouldHaveLength, 1)
			So(result[0].Resource, ShouldEqual, mcp)
			So(*result[0].Description, ShouldEqual, desc)
		})
	})
}
