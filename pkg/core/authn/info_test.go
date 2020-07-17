package authn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/authn"
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
					"X-Authgear-Session-Valid": []string{"false"},
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

				r := &http.Request{Header: rw.Header()}
				ii, err := authn.ParseHeaders(r)
				So(err, ShouldBeNil)
				So(ii, ShouldResemble, i)
			})
		})
	})
}
