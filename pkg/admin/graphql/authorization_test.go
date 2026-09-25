package graphql

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func TestAuthorizedScopes(t *testing.T) {
	Convey("Authorization.authorizedScopes", t, func() {
		r1 := &model.Resource{Meta: model.Meta{ID: "r1"}}
		r1Read := "Read orders on r1"

		var resourceBatches [][]any
		resources := graphqlutil.NewDataLoader(func(ctx context.Context, keys []any) ([]any, error) {
			resourceBatches = append(resourceBatches, keys)
			byID := map[string]*model.Resource{"r1": r1, "r2": {Meta: model.Meta{ID: "r2"}}}
			out := make([]any, len(keys))
			for i, k := range keys {
				out[i] = byID[k.(string)]
			}
			return out, nil
		})
		scopes := graphqlutil.NewDataLoader(func(ctx context.Context, keys []any) ([]any, error) {
			byKey := map[loader.ResourceScopeKey]*model.Scope{
				{ResourceID: "r1", Scope: "read:orders"}: {ResourceID: "r1", Scope: "read:orders", Description: &r1Read},
			}
			out := make([]any, len(keys))
			for i, k := range keys {
				out[i] = byKey[k.(loader.ResourceScopeKey)]
			}
			return out, nil
		})
		gqlCtx := &Context{Resources: resources, ResourceScopes: scopes}
		ctx := WithContext(context.Background(), gqlCtx)

		Convey("drops scopes granted on a deleted resource", func() {
			out, err := resolveAuthorizedScopes(ctx, gqlCtx, &model.Authorization{AuthorizedScopes: []model.AuthorizedScope{
				{Scope: "openid"},
				{ResourceID: "r1", Scope: "read:orders"},
				{ResourceID: "deleted", Scope: "read:orders"},
			}})()
			So(err, ShouldBeNil)
			So(out, ShouldResemble, []model.AuthorizedScope{
				{Scope: "openid"},
				{ResourceID: "r1", Scope: "read:orders"},
			})
		})

		Convey("batches resource loads across authorizations", func() {
			a := resolveAuthorizedScopes(ctx, gqlCtx, &model.Authorization{AuthorizedScopes: []model.AuthorizedScope{
				{ResourceID: "r1", Scope: "read:orders"},
			}})
			b := resolveAuthorizedScopes(ctx, gqlCtx, &model.Authorization{AuthorizedScopes: []model.AuthorizedScope{
				{ResourceID: "r2", Scope: "read:orders"},
			}})
			_, err := a()
			So(err, ShouldBeNil)
			_, err = b()
			So(err, ShouldBeNil)
			So(resourceBatches, ShouldResemble, [][]any{{"r1", "r2"}})
		})

		Convey("returns an empty list for an authorization with no scopes", func() {
			out, err := resolveAuthorizedScopes(ctx, gqlCtx, &model.Authorization{})()
			So(err, ShouldBeNil)
			So(out, ShouldResemble, []model.AuthorizedScope{})
		})

		resolve := func(field string, source model.AuthorizedScope) any {
			value, err := authorizedScopeType.Fields()[field].Resolve(graphql.ResolveParams{Context: ctx, Source: source})
			So(err, ShouldBeNil)
			if thunk, ok := value.(func() (any, error)); ok {
				value, err = thunk()
				So(err, ShouldBeNil)
			}
			return value
		}

		Convey("resolves the resource and the description of the scope on it", func() {
			s := model.AuthorizedScope{ResourceID: "r1", Scope: "read:orders"}
			So(resolve("resource", s), ShouldEqual, r1)
			So(resolve("description", s), ShouldEqual, &r1Read)

			So(resolve("description", model.AuthorizedScope{ResourceID: "r2", Scope: "read:orders"}), ShouldBeNil)
		})

		Convey("resolves neither for a project-level scope", func() {
			s := model.AuthorizedScope{Scope: "openid"}
			So(resolve("resource", s), ShouldBeNil)
			So(resolve("description", s), ShouldBeNil)
		})
	})
}
