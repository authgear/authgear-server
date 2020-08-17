package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

func TestInfo(t *testing.T) {
	Convey("Info", t, func() {
		Convey("should write to HTTP headers correctly", func() {
			rw := httptest.NewRecorder()

			Convey("invalid auth", func() {
				var i = &session.Info{
					IsValid: false,
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Authgear-Session-Valid": []string{"false"},
				})
			})

			Convey("valid auth", func() {
				var i = &session.Info{
					IsValid:       true,
					UserID:        "user-id",
					UserAnonymous: true,
					UserVerified:  true,
					SessionACR:    "http://schemas.openid.net/pape/policies/2007/06/multi-factor",
					SessionAMR:    []string{"pwd", "mfa", "otp"},
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":  []string{"true"},
					"X-Authgear-User-Id":        []string{"user-id"},
					"X-Authgear-User-Anonymous": []string{"true"},
					"X-Authgear-User-Verified":  []string{"true"},
					"X-Authgear-Session-Acr":    []string{"http://schemas.openid.net/pape/policies/2007/06/multi-factor"},
					"X-Authgear-Session-Amr":    []string{"pwd mfa otp"},
				})
			})
		})
	})
}
