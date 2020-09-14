package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
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
			s := &idpsession.IDPSession{
				ID: "session-id",
				Attrs: session.Attrs{
					UserID: "user-id",
				},
			}
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))

			Convey("for normal user", func() {
				userIdentities := []*identity.Info{
					{Type: authn.IdentityTypeLoginID},
				}
				identities.EXPECT().ListByUser("user-id").Return(userIdentities, nil)
				verificationService.EXPECT().IsUserVerified(userIdentities).Return(true, nil)
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
				verificationService.EXPECT().IsUserVerified(userIdentities).Return(false, nil)
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
			r = r.WithContext(session.WithInvalidSession(r.Context()))
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
