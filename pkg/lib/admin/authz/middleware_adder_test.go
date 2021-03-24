package authz_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"
	. "github.com/smartystreets/goconvey/convey"

	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func TestMiddleware(t *testing.T) {
	Convey("Middleware", t, func() {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("good"))
		})

		Convey("no auth", func() {
			m := adminauthz.Middleware{
				Logger: adminauthz.Logger{
					log.Null,
				},
				Auth: config.AdminAPIAuthNone,
			}

			r, _ := http.NewRequest("GET", "/", nil)
			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(recorder.Body.String(), ShouldEqual, "good")
		})

		Convey("jwt auth success", func() {
			// nolint:gosec
			privKey, err := rsa.GenerateKey(rand.Reader, 512)
			So(err, ShouldBeNil)

			jwkKey, err := jwk.New(privKey)
			So(err, ShouldBeNil)
			_ = jwkKey.Set("kid", "mykey")

			set := jwk.NewSet()
			_ = set.Add(jwkKey)

			m := adminauthz.Middleware{
				Logger: adminauthz.Logger{
					log.Null,
				},
				Auth:  config.AdminAPIAuthJWT,
				AppID: "app-id",
				AuthKey: &config.AdminAPIAuthKey{
					Set: set,
				},
				Clock: clock.NewMockClock(),
			}

			r, _ := http.NewRequest("GET", "/", nil)

			adder := adminauthz.Adder{
				Clock: m.Clock,
			}

			err = adder.AddAuthz(m.Auth, m.AppID, m.AuthKey, r.Header)
			So(err, ShouldBeNil)

			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(recorder.Body.String(), ShouldEqual, "good")
		})

		Convey("jwt auth failure", func() {
			// nolint:gosec
			privKey, err := rsa.GenerateKey(rand.Reader, 512)
			So(err, ShouldBeNil)

			jwkKey, err := jwk.New(privKey)
			So(err, ShouldBeNil)
			_ = jwkKey.Set("kid", "mykey")

			set := jwk.NewSet()
			_ = set.Add(jwkKey)

			m := adminauthz.Middleware{
				Logger: adminauthz.Logger{
					log.Null,
				},
				Auth:  config.AdminAPIAuthJWT,
				AppID: "app-id",
				AuthKey: &config.AdminAPIAuthKey{
					Set: set,
				},
				Clock: clock.NewMockClock(),
			}

			r, _ := http.NewRequest("GET", "/", nil)

			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(recorder.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}
