package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
)

type stubOAuthClientResolver struct {
	clients map[string]*config.OAuthClientConfig
}

func (r *stubOAuthClientResolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	return r.clients[clientID]
}

func TestResolveHandler(t *testing.T) {
	Convey("/session/resolve", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		database := &db.MockHandle{}
		userInfoService := NewMockUserInfoService(ctrl)
		h := &ResolveHandler{
			Database:        database,
			UserInfoService: userInfoService,
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
				userInfoService.EXPECT().GetUserInfoGreatest(r.Context(), "user-id").Return(
					&userinfo.UserInfo{
						User: &model.User{
							IsAnonymous:       false,
							IsVerified:        true,
							CanReauthenticate: true,
						},
						EffectiveRoleKeys: []string{},
					},
					nil,
				)
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
				userInfoService.EXPECT().GetUserInfoGreatest(r.Context(), "user-id").Return(
					&userinfo.UserInfo{
						User: &model.User{
							IsAnonymous:       true,
							IsVerified:        false,
							CanReauthenticate: false,
						},
						EffectiveRoleKeys: []string{},
					},
					nil,
				)

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

func TestResolveHandlerThirdPartyOpaqueTokenGate(t *testing.T) {
	Convey("/resolve rejects every third-party client access token, regardless of token shape", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		database := &db.MockHandle{}
		userInfoService := NewMockUserInfoService(ctrl)

		thirdPartyClient := &config.OAuthClientConfig{
			ClientID:        "third-party-client",
			ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp,
		}
		firstPartyClient := &config.OAuthClientConfig{
			ClientID:        "first-party-client",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		clientResolver := &stubOAuthClientResolver{
			clients: map[string]*config.OAuthClientConfig{
				"third-party-client": thirdPartyClient,
				"first-party-client": firstPartyClient,
			},
		}

		h := &ResolveHandler{
			Database:            database,
			UserInfoService:     userInfoService,
			OAuthClientResolver: clientResolver,
		}

		offlineGrantSession := func(clientID string) *oauth.OfflineGrantSession {
			return &oauth.OfflineGrantSession{
				OfflineGrant: &oauth.OfflineGrant{
					Attrs: session.Attrs{UserID: "user-id"},
				},
				ClientID: clientID,
			}
		}

		expectUserInfoOnce := func(ctx context.Context) {
			userInfoService.EXPECT().GetUserInfoGreatest(ctx, "user-id").Return(
				&userinfo.UserInfo{
					User:              &model.User{IsAnonymous: false, IsVerified: true, CanReauthenticate: true},
					EffectiveRoleKeys: []string{},
				},
				nil,
			)
		}

		Convey("third-party client, opaque token: invalid", func() {
			s := offlineGrantSession("third-party-client")
			s.TokenType = session.TokenTypeOpaque
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Authgear-Session-Valid": []string{"false"},
			})
		})

		Convey("third-party client, resource-bound JWT: invalid", func() {
			s := offlineGrantSession("third-party-client")
			s.TokenType = session.TokenTypeJWT
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Authgear-Session-Valid": []string{"false"},
			})
		})

		Convey("first-party client, JWT (project-endpoint default or resource-bound): succeeds, unchanged", func() {
			s := offlineGrantSession("first-party-client")
			s.TokenType = session.TokenTypeJWT
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))
			expectUserInfoOnce(r.Context())

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("X-Authgear-Session-Valid"), ShouldEqual, "true")
		})

		Convey("first-party client, opaque token: succeeds, unchanged", func() {
			s := offlineGrantSession("first-party-client")
			s.TokenType = session.TokenTypeOpaque
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))
			expectUserInfoOnce(r.Context())

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("X-Authgear-Session-Valid"), ShouldEqual, "true")
		})

		Convey("third-party client, app session token cookie: invalid, even though this combination cannot occur in practice", func() {
			s := offlineGrantSession("third-party-client")
			s.TokenType = session.TokenTypeAppSession
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header, ShouldResemble, http.Header{
				"X-Authgear-Session-Valid": []string{"false"},
			})
		})

		Convey("first-party client, app session token cookie: succeeds, unchanged", func() {
			s := offlineGrantSession("first-party-client")
			s.TokenType = session.TokenTypeAppSession
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))
			expectUserInfoOnce(r.Context())

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("X-Authgear-Session-Valid"), ShouldEqual, "true")
		})

		Convey("IDP session: always succeeds regardless of TokenType", func() {
			s := &idpsession.IDPSession{
				ID:    "session-id",
				Attrs: session.Attrs{UserID: "user-id"},
			}
			r, _ := http.NewRequest("POST", "/", nil)
			r = r.WithContext(session.WithSession(r.Context(), s))
			expectUserInfoOnce(r.Context())

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, r)

			resp := rw.Result()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("X-Authgear-Session-Valid"), ShouldEqual, "true")
		})
	})
}
