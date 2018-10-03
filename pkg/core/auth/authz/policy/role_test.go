package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRolePolicy(t *testing.T) {
	Convey("should return error if auth context has no auth info", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{}

		err := NewAllowRole("roleA").IsAllowed(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should return error if specified allowed role does not exist", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:    "ID",
				Roles: []string{"roleA", "roleB"},
			},
		}

		err := NewAllowRole("admin").IsAllowed(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should pass if specified allowed role exists", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:    "ID",
				Roles: []string{"roleA", "roleB"},
			},
		}

		err := NewAllowRole("roleA").IsAllowed(req, ctx)
		So(err, ShouldBeEmpty)
	})

	Convey("should return error if specified denied role exists", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:    "ID",
				Roles: []string{"roleA", "roleB"},
			},
		}

		err := NewDenyRole("roleA").IsAllowed(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should pass if specified denied role does not exist", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:    "ID",
				Roles: []string{"roleA", "roleB"},
			},
		}

		err := NewDenyRole("admin").IsAllowed(req, ctx)
		So(err, ShouldBeEmpty)
	})
}
