package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authapi "github.com/authgear/authgear-server/pkg/auth/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

func TestLoginLinkV1InspectHandler(t *testing.T) {
	Convey("LoginLinkV1InspectHandler", t, func() {
		service := &stubLoginLinkService{}
		h := &LoginLinkV1InspectHandler{
			Controller: &authapi.APIController{LoginLinks: service},
		}

		Convey("inspects the code without approving it", func() {
			status, resp := serveLoginLinkV1(h, "application/json", `{"code": "CODE"}`)
			So(status, ShouldEqual, 200)
			So(resp, ShouldResemble, map[string]any{"result": map[string]any{}})
			So(service.inspects, ShouldResemble, []string{"CODE"})
			So(service.approves, ShouldBeEmpty)
		})

		Convey("returns the service error", func() {
			service.err = otp.ErrCodeNotFound

			status, resp := serveLoginLinkV1(h, "application/json", `{"code": "CODE"}`)
			So(status, ShouldEqual, 403)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "InvalidOTPCode")
		})

		Convey("requires code", func() {
			status, resp := serveLoginLinkV1(h, "application/json", `{}`)
			So(status, ShouldEqual, 400)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "ValidationFailed")
			So(service.inspects, ShouldBeEmpty)
		})
	})
}
