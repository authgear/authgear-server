package authz_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwk"
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
			privKey, err := rsa.GenerateKey(rand.Reader, 2048)
			So(err, ShouldBeNil)

			jwkKey, err := jwk.FromRaw(privKey)
			So(err, ShouldBeNil)
			_ = jwkKey.Set(jwk.KeyIDKey, "mykey")
			_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
			_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

			set := jwk.NewSet()
			_ = set.AddKey(jwkKey)

			m := adminauthz.Middleware{
				Logger: adminauthz.Logger{
					log.Null,
				},
				Auth:  config.AdminAPIAuthJWT,
				AppID: "app-id",
				AuthKey: &config.AdminAPIAuthKey{
					Set: set,
				},
				Clock: clock.NewMockClockAt("2001-09-09T01:46:40.000Z"),
			}

			r, _ := http.NewRequest("GET", "/", nil)

			adder := adminauthz.Adder{
				Clock: m.Clock,
			}

			err = adder.AddAuthz(m.Auth, m.AppID, m.AuthKey, nil, r.Header)
			So(err, ShouldBeNil)

			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(recorder.Body.String(), ShouldEqual, "good")
		})

		Convey("jwt auth failure", func() {
			privKey, err := rsa.GenerateKey(rand.Reader, 2048)
			So(err, ShouldBeNil)

			jwkKey, err := jwk.FromRaw(privKey)
			So(err, ShouldBeNil)
			_ = jwkKey.Set(jwk.KeyIDKey, "mykey")
			_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
			_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

			set := jwk.NewSet()
			_ = set.AddKey(jwkKey)

			m := adminauthz.Middleware{
				Logger: adminauthz.Logger{
					log.Null,
				},
				Auth:  config.AdminAPIAuthJWT,
				AppID: "app-id",
				AuthKey: &config.AdminAPIAuthKey{
					Set: set,
				},
				Clock: clock.NewMockClockAt("2001-09-09T01:46:40.000Z"),
			}

			r, _ := http.NewRequest("GET", "/", nil)

			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(recorder.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}
