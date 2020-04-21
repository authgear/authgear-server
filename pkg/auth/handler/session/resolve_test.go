package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		h := &ResolveHandler{
			TimeProvider: &time.MockProvider{},
		}

		Convey("should attach headers for valid sessions", func() {
			u := &authn.UserInfo{
				ID:         "user-id",
				IsDisabled: false,
				IsVerified: true,
			}
			s := &session.IDPSession{
				ID: "session-id",
				Attrs: authn.Attrs{
					IdentityType:   "password",
					IdentityClaims: map[string]interface{}{},
				},
			}
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(authn.WithAuthn(r.Context(), s, u))
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Skygear-Session-Valid":           []string{"true"},
				"X-Skygear-User-Id":                 []string{"user-id"},
				"X-Skygear-User-Verified":           []string{"true"},
				"X-Skygear-User-Disabled":           []string{"false"},
				"X-Skygear-Session-Identity-Type":   []string{"password"},
				"X-Skygear-Session-Identity-Claims": []string{"e30"},
				"X-Skygear-Session-Acr":             []string{""},
				"X-Skygear-Session-Amr":             []string{""},
				"X-Skygear-Is-Master-Key":           []string{"false"},
			})
		})

		Convey("should attach headers for invalid sessions", func() {
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(authn.WithInvalidAuthn(r.Context()))
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Skygear-Session-Valid": []string{"false"},
				"X-Skygear-Is-Master-Key": []string{"false"},
			})
		})

		Convey("should not attach session headers if no resolved session", func() {
			r, _ := http.NewRequest("POST", "/", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Skygear-Is-Master-Key": []string{"false"},
			})
		})

		Convey("should add master key header", func() {
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(coreauth.WithAccessKey(r.Context(), coreauth.AccessKey{
				IsMasterKey: true,
			}))
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Skygear-Is-Master-Key": []string{"true"},
			})
		})
	})
}
