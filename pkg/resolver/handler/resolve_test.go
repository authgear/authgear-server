package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identities := NewMockIdentityService(ctrl)
		verificationService := NewMockVerificationService(ctrl)
		database := &db.MockHandle{}
		user := NewMockUserProvider(ctrl)
		roleAndGroup := NewMockRolesAndGroupsProvider(ctrl)
		h := &ResolveHandler{
			Database:       database,
			Identities:     identities,
			Verification:   verificationService,
			Users:          user,
			RolesAndGroups: roleAndGroup,
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
					{Type: model.IdentityTypeLoginID},
				}
				identities.EXPECT().ListByUser(r.Context(), "user-id").Return(userIdentities, nil)
				verificationService.EXPECT().IsUserVerified(r.Context(), userIdentities).Return(true, nil)
				userInfo := model.User{
					CanReauthenticate: true,
				}
				user.EXPECT().Get(r.Context(), "user-id", accesscontrol.RoleGreatest).Return(&userInfo, nil)
				roles := []*model.Role{}
				roleAndGroup.EXPECT().ListEffectiveRolesByUserID(r.Context(), "user-id").Return(roles, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":           []string{"true"},
					"X-Authgear-User-Id":                 []string{"user-id"},
					"X-Authgear-User-Verified":           []string{"true"},
					"X-Authgear-User-Anonymous":          []string{"false"},
					"X-Authgear-Session-Amr":             []string{""},
					"X-Authgear-User-Can-Reauthenticate": []string{"true"},
				})
			})

			Convey("for anonymous user", func() {
				userIdentities := []*identity.Info{
					{Type: model.IdentityTypeAnonymous},
					{Type: model.IdentityTypeLoginID},
				}
				identities.EXPECT().ListByUser(r.Context(), "user-id").Return(userIdentities, nil)
				verificationService.EXPECT().IsUserVerified(r.Context(), userIdentities).Return(false, nil)
				userInfo := model.User{
					CanReauthenticate: false,
				}
				user.EXPECT().Get(r.Context(), "user-id", accesscontrol.RoleGreatest).Return(&userInfo, nil)
				roles := []*model.Role{}
				roleAndGroup.EXPECT().ListEffectiveRolesByUserID(r.Context(), "user-id").Return(roles, nil)
				rw := httptest.NewRecorder()
				h.ServeHTTP(rw, r)

				resp := rw.Result()
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header, ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":           []string{"true"},
					"X-Authgear-User-Id":                 []string{"user-id"},
					"X-Authgear-User-Anonymous":          []string{"true"},
					"X-Authgear-User-Verified":           []string{"false"},
					"X-Authgear-Session-Amr":             []string{""},
					"X-Authgear-User-Can-Reauthenticate": []string{"false"},
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
