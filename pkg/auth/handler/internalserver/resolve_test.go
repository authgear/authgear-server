package internalserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identities := NewMockIdentityService(ctrl)
		verificationService := NewMockVerificationService(ctrl)
		h := &ResolveHandler{
			Identities:   identities,
			Verification: verificationService,
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
				userIdentities := []*identity.Info{
					{Type: authn.IdentityTypeLoginID},
				}
				identities.EXPECT().ListByUser("user-id").Return(userIdentities, nil)
				verificationService.EXPECT().IsUserVerified(userIdentities, "user-id").Return(true, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":  []string{"true"},
					"X-Authgear-User-Id":        []string{"user-id"},
					"X-Authgear-User-Verified":  []string{"true"},
					"X-Authgear-User-Anonymous": []string{"false"},
					"X-Authgear-Session-Acr":    []string{""},
					"X-Authgear-Session-Amr":    []string{""},
				})
			})

			Convey("for anonymous user", func() {
				userIdentities := []*identity.Info{
					{Type: authn.IdentityTypeAnonymous},
					{Type: authn.IdentityTypeLoginID},
				}
				identities.EXPECT().ListByUser("user-id").Return(userIdentities, nil)
				verificationService.EXPECT().IsUserVerified(userIdentities, "user-id").Return(false, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":  []string{"true"},
					"X-Authgear-User-Id":        []string{"user-id"},
					"X-Authgear-User-Anonymous": []string{"true"},
					"X-Authgear-User-Verified":  []string{"false"},
					"X-Authgear-Session-Acr":    []string{""},
					"X-Authgear-Session-Amr":    []string{""},
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
				"X-Authgear-Session-Valid": []string{"false"},
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
