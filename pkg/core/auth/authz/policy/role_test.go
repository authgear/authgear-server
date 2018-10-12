package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRolePolicy(t *testing.T) {
	Convey("Test RolePolicy", t, func() {
		Convey("should return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := NewAllowRole("roleA").IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should return error if specified allowed role does not exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:    "ID",
					Roles: []string{"roleA", "roleB"},
				},
			}

			err := NewAllowRole("admin").IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should pass if specified allowed role exists", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:    "ID",
					Roles: []string{"roleA", "roleB"},
				},
			}

			err := NewAllowRole("roleA").IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if specified denied role exists", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:    "ID",
					Roles: []string{"roleA", "roleB"},
				},
			}

			err := NewDenyRole("roleA").IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should pass if specified denied role does not exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:    "ID",
					Roles: []string{"roleA", "roleB"},
				},
			}

			err := NewDenyRole("admin").IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

	})
}
