package handler

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func TestLogoutHandler(t *testing.T) {
	Convey("Test LogoutRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := LogoutRequestPayload{
				AccessToken: "test_token",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate missing access token", func() {
			payload := LogoutRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.AccessTokenNotAccepted)
		})
	})

	Convey("Test LogoutHandler", t, func() {
		h := &LogoutHandler{}
		h.TokenStore = authtoken.NewJWTStore("myApp", "secret", 0)

		Convey("Test LogoutHandler", func() {
			token := "test_token"
			payload := LogoutRequestPayload{
				AccessToken: token,
			}
			resp, err := h.Handle(payload)
			So(resp, ShouldEqual, "OK")
			So(err, ShouldBeNil)
		})
	})
}
