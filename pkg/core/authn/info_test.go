package authn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
					IsValid:                        true,
					UserID:                         "user-id",
					UserVerified:                   true,
					UserDisabled:                   false,
					SessionIdentityID:              "principal-id",
					SessionIdentityType:            "password",
					SessionIdentityUpdatedAt:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					SessionAuthenticatorID:         "authenticator-id",
					SessionAuthenticatorType:       "oob",
					SessionAuthenticatorOOBChannel: "email",
					SessionAuthenticatorUpdatedAt:  nil,
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Skygear-Session-Valid":                     []string{"true"},
					"X-Skygear-User-Id":                           []string{"user-id"},
					"X-Skygear-User-Verified":                     []string{"true"},
					"X-Skygear-User-Disabled":                     []string{"false"},
					"X-Skygear-Session-Identity-Id":               []string{"principal-id"},
					"X-Skygear-Session-Identity-Type":             []string{"password"},
					"X-Skygear-Session-Identity-Updated-At":       []string{"2020-01-01T00:00:00Z"},
					"X-Skygear-Session-Authenticator-Id":          []string{"authenticator-id"},
					"X-Skygear-Session-Authenticator-Type":        []string{"oob"},
					"X-Skygear-Session-Authenticator-Oob-Channel": []string{"email"},
				})

				r := &http.Request{Header: rw.Header()}
				ii, err := authn.ParseHeaders(r)
				So(err, ShouldBeNil)
				So(ii, ShouldResemble, i)
			})
		})
	})
}
