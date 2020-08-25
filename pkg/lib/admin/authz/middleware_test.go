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
			w.Write([]byte("good"))
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

			So(string(recorder.Body.Bytes()), ShouldEqual, "good")
		})

		Convey("jwt auth", func() {
			// nolint:gosec
			privKey, err := rsa.GenerateKey(rand.Reader, 512)
			So(err, ShouldBeNil)

			jwkKey, err := jwk.New(privKey)
			So(err, ShouldBeNil)
			jwkKey.Set("kid", "mykey")

			set := jwk.Set{
				Keys: []jwk.Key{jwkKey},
			}

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

			authzAdder := adminauthz.AuthzAdder{
				Clock: m.Clock,
			}

			err = authzAdder.AddAuthz(m.Auth, m.AppID, m.AuthKey, r.Header)
			So(err, ShouldBeNil)

			recorder := httptest.NewRecorder()
			handler := m.Handle(h)
			handler.ServeHTTP(recorder, r)

			So(string(recorder.Body.Bytes()), ShouldEqual, "good")
		})
	})
}
