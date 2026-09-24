package authflowv2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResourcePermissions(t *testing.T) {
	Convey("resourcePermissions", t, func() {
		descriptions := map[string]string{
			"read:tools":    "Show all tools",
			"execute:tools": "",
		}

		Convey("skips identity and protocol scopes", func() {
			got := resourcePermissions([]string{
				"openid",
				"offline_access",
				"profile",
				"email",
				"phone",
				"address",
				"https://authgear.com/scopes/full-userinfo",
				"device_sso",
				"https://authgear.com/scopes/pre-authenticated-url",
				"https://authgear.com/scopes/full-access",
			}, descriptions)
			So(got, ShouldBeEmpty)
		})

		Convey("uses the description when one is configured", func() {
			got := resourcePermissions([]string{"read:tools"}, descriptions)
			So(got, ShouldResemble, []AuthorizationPermission{
				{Scope: "read:tools", DisplayText: "Show all tools"},
			})
		})

		Convey("falls back to the raw scope name when the description is empty or the scope is unknown", func() {
			got := resourcePermissions([]string{"execute:tools", "unknown:scope"}, descriptions)
			So(got, ShouldResemble, []AuthorizationPermission{
				{Scope: "execute:tools", DisplayText: "execute:tools"},
				{Scope: "unknown:scope", DisplayText: "unknown:scope"},
			})
		})

		Convey("keeps grant order and tolerates a nil description map", func() {
			got := resourcePermissions([]string{"openid", "b:scope", "a:scope"}, nil)
			So(got, ShouldResemble, []AuthorizationPermission{
				{Scope: "b:scope", DisplayText: "b:scope"},
				{Scope: "a:scope", DisplayText: "a:scope"},
			})
		})
	})
}
