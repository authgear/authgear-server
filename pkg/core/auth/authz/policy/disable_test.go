package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyDisabledUser(t *testing.T) {
	Convey("should return error if auth context has no auth info", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{}

		err := DenyDisabledUser(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should return error if user is disabled", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:       "ID",
				Disabled: true,
			},
		}

		err := DenyDisabledUser(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should pass if user is not disabled", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AuthInfo: &authinfo.AuthInfo{
				ID:       "ID",
				Disabled: false,
			},
		}

		err := DenyDisabledUser(req, ctx)
		So(err, ShouldBeEmpty)
	})
}
