package pq

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type fakeAuthzRow struct {
	scopesByResources []byte
}

func (r fakeAuthzRow) Scan(dest ...any) error {
	*dest[0].(*string) = "authz"
	*dest[1].(*string) = "app"
	*dest[2].(*string) = "client"
	*dest[3].(*string) = "user"
	*dest[4].(*time.Time) = time.Time{}
	*dest[5].(*time.Time) = time.Time{}
	*dest[6].(*[]byte) = r.scopesByResources
	return nil
}

func TestAuthorizationStoreScan(t *testing.T) {
	Convey("AuthorizationStore.scanAuthz reads scopes_by_resources", t, func() {
		s := &AuthorizationStore{}
		authz, err := s.scanAuthz(fakeAuthzRow{
			scopesByResources: []byte(`{"authgear":["openid"],"r1":["read:orders"]}`),
		})
		So(err, ShouldBeNil)
		So(authz.ScopesByResources, ShouldResemble, map[string][]string{
			oauth.AuthorizationProjectScopesKey: {"openid"},
			"r1":                                {"read:orders"},
		})
	})
}

func TestMarshalScopes(t *testing.T) {
	Convey("marshalScopes", t, func() {
		Convey("writes only project-level scopes to scopes", func() {
			scopes, scopesByResources, err := marshalScopes(&oauth.Authorization{ScopesByResources: map[string][]string{
				oauth.AuthorizationProjectScopesKey: {"openid"},
				"r1":                                {"read:orders", "write:orders"},
				"r2":                                {"read:orders"},
			}})
			So(err, ShouldBeNil)
			So(string(scopes), ShouldEqual, `["openid"]`)
			So(string(scopesByResources), ShouldEqual, `{"authgear":["openid"],"r1":["read:orders","write:orders"],"r2":["read:orders"]}`)
		})

		Convey("writes empty values for no scopes", func() {
			scopes, scopesByResources, err := marshalScopes(&oauth.Authorization{})
			So(err, ShouldBeNil)
			So(string(scopes), ShouldEqual, `[]`)
			So(string(scopesByResources), ShouldEqual, `{}`)
		})
	})
}
