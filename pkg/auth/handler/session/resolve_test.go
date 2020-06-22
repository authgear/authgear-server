package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		anonymousProvider := NewMockAnonymousIdentityProvider(ctrl)
		h := &ResolveHandler{
			Anonymous: anonymousProvider,
		}

		Convey("should attach headers for valid sessions", func() {
			u := &authn.UserInfo{
				ID: "user-id",
			}
			s := &session.IDPSession{
				ID:    "session-id",
				Attrs: authn.Attrs{},
			}
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(authn.WithAuthn(r.Context(), s, u))

			Convey("for normal user", func() {
				anonymousProvider.EXPECT().List("user-id").Return([]*anonymous.Identity{}, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Skygear-Session-Valid":  []string{"true"},
					"X-Skygear-User-Id":        []string{"user-id"},
					"X-Skygear-User-Anonymous": []string{"false"},
					"X-Skygear-Session-Acr":    []string{""},
					"X-Skygear-Session-Amr":    []string{""},
				})
			})

			Convey("for anonymous user", func() {
				anonymousProvider.EXPECT().List("user-id").Return([]*anonymous.Identity{
					{ID: "anonymous-identity"},
				}, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Skygear-Session-Valid":  []string{"true"},
					"X-Skygear-User-Id":        []string{"user-id"},
					"X-Skygear-User-Anonymous": []string{"true"},
					"X-Skygear-Session-Acr":    []string{""},
					"X-Skygear-Session-Amr":    []string{""},
				})
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
			})
		})

		Convey("should not attach session headers if no resolved session", func() {
			r, _ := http.NewRequest("POST", "/", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{})
		})
	})
}
