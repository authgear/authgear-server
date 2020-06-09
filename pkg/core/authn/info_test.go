package authn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthnInfo(t *testing.T) {
	Convey("AuthnInfo", t, func() {
		Convey("should round-trip to/from HTTP headers correctly", func() {
			rw := httptest.NewRecorder()

			Convey("no auth", func() {
				var i *authn.Info

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{})

				r := &http.Request{Header: rw.Header()}
				ii, err := authn.ParseHeaders(r)
				So(err, ShouldBeNil)
				So(ii, ShouldResemble, i)
			})

			Convey("invalid auth", func() {
				var i *authn.Info = &authn.Info{
					IsValid: false,
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Skygear-Session-Valid": []string{"false"},
				})

				r := &http.Request{Header: rw.Header()}
				ii, err := authn.ParseHeaders(r)
				So(err, ShouldBeNil)
				So(ii, ShouldResemble, i)
			})

			Convey("valid auth", func() {
				var i *authn.Info = &authn.Info{
					IsValid:       true,
					UserID:        "user-id",
					UserDisabled:  false,
					UserAnonymous: true,
					SessionACR:    "http://schemas.openid.net/pape/policies/2007/06/multi-factor",
					SessionAMR:    []string{"pwd", "mfa", "otp"},
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Skygear-Session-Valid":  []string{"true"},
					"X-Skygear-User-Id":        []string{"user-id"},
					"X-Skygear-User-Disabled":  []string{"false"},
					"X-Skygear-User-Anonymous": []string{"true"},
					"X-Skygear-Session-Acr":    []string{"http://schemas.openid.net/pape/policies/2007/06/multi-factor"},
					"X-Skygear-Session-Amr":    []string{"pwd mfa otp"},
				})

				r := &http.Request{Header: rw.Header()}
				ii, err := authn.ParseHeaders(r)
				So(err, ShouldBeNil)
				So(ii, ShouldResemble, i)
			})
		})
	})
}
