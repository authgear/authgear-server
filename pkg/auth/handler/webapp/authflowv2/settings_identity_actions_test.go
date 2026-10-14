package authflowv2

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSettingsListURLMakers(t *testing.T) {
	Convey("SettingsListURLMakers", t, func() {
		makeURL := func(path string, pairs ...string) string { return path }
		loginID := func(id string, t model.LoginIDKeyType) *identity.Info {
			return &identity.Info{
				ID:      id,
				Type:    model.IdentityTypeLoginID,
				LoginID: &identity.LoginID{LoginIDType: t},
			}
		}
		oauthWithEmail := &identity.Info{
			Type:  model.IdentityTypeOAuth,
			OAuth: &identity.OAuth{Claims: map[string]any{stdattrs.Email: "a@example.com"}},
		}

		Convey("skips the list when there is at most one item", func() {
			m := NewSettingsListURLMakers(makeURL, []*identity.Info{loginID("1", model.LoginIDKeyTypePhone)})
			So(m.MakeListPhoneURL("/skip"), ShouldEqual, "/skip")
			// An empty list has nothing to show either.
			So(m.MakeListEmailURL("/skip"), ShouldEqual, "/skip")
		})

		Convey("keeps the list when there are several items", func() {
			m := NewSettingsListURLMakers(makeURL, []*identity.Info{
				loginID("1", model.LoginIDKeyTypeEmail),
				loginID("2", model.LoginIDKeyTypeEmail),
			})
			So(m.MakeListEmailURL("/skip"), ShouldEqual, "/settings/identity/email")
		})

		Convey("counts an OAuth identity with an email claim", func() {
			m := NewSettingsListURLMakers(makeURL, []*identity.Info{
				loginID("1", model.LoginIDKeyTypeEmail),
				oauthWithEmail,
			})
			So(m.MakeListEmailURL("/skip"), ShouldEqual, "/settings/identity/email")
		})
	})
}
