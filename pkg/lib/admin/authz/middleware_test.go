package authz_test

import (
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
				Logger: adminauthz.AuthzLogger{
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

		SkipConvey("jwt auth", func() {
			key, err := jwk.New([]byte("secret"))
			So(err, ShouldBeNil)

			set := jwk.Set{
				Keys: []jwk.Key{key},
			}

			m := adminauthz.Middleware{
				Logger: adminauthz.AuthzLogger{
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

			So(string(recorder.Body.Bytes()), ShouldEqual, "good")
		})
	})
}
