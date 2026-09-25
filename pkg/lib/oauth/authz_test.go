package oauth_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func TestAuthorization(t *testing.T) {
	Convey("Authorization", t, func() {
		Convey("IsAuthorized", func() {
			authz := &oauth.Authorization{ScopesByResources: map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid", "email"},
				"r1":                                {"read:orders"},
			}}
			So(authz.IsAuthorized("", []string{"openid"}), ShouldBeTrue)
			So(authz.IsAuthorized("", []string{"openid", "email"}), ShouldBeTrue)
			So(authz.IsAuthorized("", []string{"profile"}), ShouldBeFalse)
			So(authz.IsAuthorized("r1", []string{"openid", "read:orders"}), ShouldBeTrue)
			So(authz.IsAuthorized("r1", []string{"read:orders", "write:orders"}), ShouldBeFalse)

			Convey("a scope granted on one resource is not granted on another", func() {
				So(authz.IsAuthorized("r2", []string{"read:orders"}), ShouldBeFalse)
				So(authz.IsAuthorized("r2", []string{"openid"}), ShouldBeTrue)
			})

			Convey("a resource scope without a resource is ignored", func() {
				So(authz.IsAuthorized("", []string{"openid", "write:orders"}), ShouldBeTrue)
				So(authz.IsAuthorized("", []string{"profile", "write:orders"}), ShouldBeFalse)
			})
		})

		Convey("WithScopesAdded", func() {
			authz := &oauth.Authorization{ScopesByResources: map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r1":                                {"read:orders"},
			}}

			added := authz.WithScopesAdded("r2", []string{"openid", "email", "read:orders"})
			So(added.ScopesByResources, ShouldResemble, map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid", "email"},
				"r1":                                {"read:orders"},
				"r2":                                {"read:orders"},
			})
			So(authz.ScopesByResources["r2"], ShouldBeNil)

			Convey("a resource scope without a resource is ignored", func() {
				added := (&oauth.Authorization{}).WithScopesAdded("", []string{"openid", "read:orders"})
				So(added.ScopesByResources, ShouldResemble, map[string][]string{
					oauth.AuthorizationProjectScopesKey: {"openid"},
				})
			})
		})

		Convey("WithResourcesRemoved", func() {
			authz := &oauth.Authorization{ScopesByResources: map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r1":                                {"read:orders"},
				"r2":                                {"read:orders"},
			}}
			So(authz.WithResourcesRemoved([]string{"r1"}).ScopesByResources, ShouldResemble, map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r2":                                {"read:orders"},
			})
		})

		Convey("AllScopes", func() {
			authz := &oauth.Authorization{ScopesByResources: map[string][]string{
				"r2":                                {"read:orders", "write:orders"},
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r1":                                {"read:orders"},
			}}
			So(authz.AllScopes(), ShouldResemble, []string{"openid", "read:orders", "write:orders"})
		})

		Convey("ToAPIModel", func() {
			authz := &oauth.Authorization{ScopesByResources: map[string][]string{
				"r2":                                {"read:orders"},
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r1":                                {"read:orders"},
			}}
			m := authz.ToAPIModel()
			So(m.Scopes, ShouldResemble, []string{"openid"})
			So(m.AuthorizedScopes, ShouldResemble, []model.AuthorizedScope{
				{Scope: "openid"},
				{ResourceID: "r1", Scope: "read:orders"},
				{ResourceID: "r2", Scope: "read:orders"},
			})

			m = (&oauth.Authorization{}).ToAPIModel()
			So(m.Scopes, ShouldResemble, []string{})
			So(m.AuthorizedScopes, ShouldResemble, []model.AuthorizedScope{})
		})
	})
}
