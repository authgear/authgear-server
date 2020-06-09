package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		anonymousProvider := NewMockAnonymousIdentityProvider(ctrl)
		h := &ResolveHandler{
			TimeProvider: &time.MockProvider{},
			Anonymous:    anonymousProvider,
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
					"X-Skygear-Is-Master-Key":  []string{"false"},
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
					"X-Skygear-Is-Master-Key":  []string{"false"},
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
