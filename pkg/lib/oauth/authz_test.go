package oauth_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func TestAuthorization(t *testing.T) {
	Convey("Authorization", t, func() {
		Convey("IsAuthorized", func() {
			authz := &oauth.Authorization{Scopes: []string{"a", "b"}}
			So(authz.IsAuthorized([]string{"a"}), ShouldBeTrue)
			So(authz.IsAuthorized([]string{"a", "b"}), ShouldBeTrue)
			So(authz.IsAuthorized([]string{"c"}), ShouldBeFalse)
			So(authz.IsAuthorized([]string{"a", "b", "c"}), ShouldBeFalse)
		})
		Convey("WithScopesAdded", func() {
			authz := &oauth.Authorization{Scopes: []string{"a", "b"}}
			So(authz.WithScopesAdded([]string{}).Scopes, ShouldResemble, []string{"a", "b"})
			So(authz.WithScopesAdded([]string{"a"}).Scopes, ShouldResemble, []string{"a", "b"})
			So(authz.WithScopesAdded([]string{"b", "c"}).Scopes, ShouldResemble, []string{"a", "b", "c"})
			So(authz.WithScopesAdded([]string{"c", "d"}).Scopes, ShouldResemble, []string{"a", "b", "c", "d"})
			authz = &oauth.Authorization{Scopes: []string{}}
			So(authz.WithScopesAdded([]string{}).Scopes, ShouldBeEmpty)
			So(authz.WithScopesAdded([]string{"a", "b"}).Scopes, ShouldResemble, []string{"a", "b"})
		})
	})
}
