package session

import (
	"testing"
	gotime "time"

	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"

	"github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
	"github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	Convey("Provider", t, func() {
		store := NewMockStore()

		timeProvider := &time.MockProvider{}
		initialTime := gotime.Date(2006, 1, 1, 0, 0, 0, 0, gotime.UTC)
		timeProvider.TimeNow = initialTime
		timeProvider.TimeNowUTC = initialTime

		authContext := authtest.NewMockContext().UseAPIAccessKey("web-app")
		clientConfigs := map[string]config.APIClientConfiguration{
			"web-app":    config.APIClientConfiguration{},
			"mobile-app": config.APIClientConfiguration{},
		}

		var provider Provider = &providerImpl{
			store:         store,
			authContext:   authContext,
			clientConfigs: clientConfigs,
			time:          timeProvider,
			rand:          corerand.InsecureRand,
		}

		Convey("creating session", func() {
			Convey("should be successful", func() {
				session, err := provider.Create("user-id", "principal-id")
				So(err, ShouldBeNil)
				So(session, ShouldResemble, &auth.Session{
					ID:                   session.ID,
					ClientID:             "web-app",
					UserID:               "user-id",
					PrincipalID:          "principal-id",
					CreatedAt:            initialTime,
					AccessedAt:           initialTime,
					AccessToken:          session.AccessToken,
					RefreshToken:         session.RefreshToken,
					AccessTokenCreatedAt: initialTime,
				})
				So(session.AccessToken, ShouldHaveLength, tokenLength+len(session.ID)+1)
			})

			Convey("should allow creating multiple sessions for same principal", func() {
				session1, err := provider.Create("user-id", "principal-id")
				So(err, ShouldBeNil)
				So(session1, ShouldResemble, &auth.Session{
					ID:                   session1.ID,
					ClientID:             "web-app",
					UserID:               "user-id",
					PrincipalID:          "principal-id",
					CreatedAt:            initialTime,
					AccessedAt:           initialTime,
					AccessToken:          session1.AccessToken,
					RefreshToken:         session1.RefreshToken,
					AccessTokenCreatedAt: initialTime,
				})

				session2, err := provider.Create("user-id", "principal-id")
				So(err, ShouldBeNil)
				So(session2, ShouldResemble, &auth.Session{
					ID:                   session2.ID,
					ClientID:             "web-app",
					UserID:               "user-id",
					PrincipalID:          "principal-id",
					CreatedAt:            initialTime,
					AccessedAt:           initialTime,
					AccessToken:          session2.AccessToken,
					RefreshToken:         session2.RefreshToken,
					AccessTokenCreatedAt: initialTime,
				})

				So(session1.ID, ShouldNotEqual, session2.ID)
			})
			Convey("should generate refresh token if enabled", func() {
				clientConfigs["web-app"] = config.APIClientConfiguration{
					RefreshTokenDisabled: false,
				}
				session, err := provider.Create("user-id", "principal-id")
				So(err, ShouldBeNil)
				So(session.RefreshToken, ShouldHaveLength, tokenLength+len(session.ID)+1)
			})
			Convey("should not generate refresh token if disabled", func() {
				clientConfigs["web-app"] = config.APIClientConfiguration{
					RefreshTokenDisabled: true,
				}
				session, err := provider.Create("user-id", "principal-id")
				So(err, ShouldBeNil)
				So(session.RefreshToken, ShouldBeEmpty)
			})
		})

		Convey("getting session", func() {
			fixtureSession := auth.Session{
				ID:                   "session-id",
				ClientID:             "web-app",
				UserID:               "user-id",
				PrincipalID:          "principal-id",
				CreatedAt:            initialTime,
				AccessedAt:           initialTime,
				AccessToken:          "session-id.access-token",
				RefreshToken:         "session-id.refresh-token",
				AccessTokenCreatedAt: initialTime,
			}
			store.Sessions[fixtureSession.ID] = fixtureSession

			Convey("should be successful using access token", func() {
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, &fixtureSession)
			})

			Convey("should be successful using refresh token", func() {
				session, err := provider.GetByToken("session-id.refresh-token", auth.SessionTokenKindRefreshToken)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, &fixtureSession)
			})

			Convey("should not mix up access & refresh token", func() {
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindRefreshToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)

				session, err = provider.GetByToken("session-id.refresh-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})

			Convey("should not match empty tokens", func() {
				Convey("for access token", func() {
					fixtureSession.AccessToken = ""
					store.Sessions[fixtureSession.ID] = fixtureSession

					session, err := provider.GetByToken("session-id.", auth.SessionTokenKindAccessToken)
					So(err, ShouldBeError, ErrSessionNotFound)
					So(session, ShouldBeNil)
				})
				Convey("for refresh token", func() {
					fixtureSession.RefreshToken = ""
					store.Sessions[fixtureSession.ID] = fixtureSession

					session, err := provider.GetByToken("session-id.", auth.SessionTokenKindRefreshToken)
					So(err, ShouldBeError, ErrSessionNotFound)
					So(session, ShouldBeNil)
				})
			})

			Convey("should reject session of other clients", func() {
				authContext.UseAPIAccessKey("mobile-app")
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})

			Convey("should reject non-existant session", func() {
				session, err := provider.GetByToken("session-id-unknown.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})

			Convey("should reject incorrect token", func() {
				session, err := provider.GetByToken("session-id.incorrect-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)

				session, err = provider.GetByToken("invalid-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
			Convey("should reject if client is disabled", func() {
				clientConfigs["web-app"] = config.APIClientConfiguration{
					Disabled: true,
				}
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
			Convey("should reject if client does not exists", func() {
				delete(clientConfigs, "web-app")
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
			Convey("should reject if session is expired", func() {
				timeProvider.AdvanceSeconds(1000000)
				session, err := provider.GetByToken("session-id.access-token", auth.SessionTokenKindAccessToken)
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
		})

		Convey("accessing session", func() {
			session := auth.Session{
				ID:                   "session-id",
				UserID:               "user-id",
				PrincipalID:          "principal-id",
				CreatedAt:            initialTime,
				AccessedAt:           initialTime,
				AccessToken:          "access-token",
				AccessTokenCreatedAt: initialTime,
			}
			timeProvider.AdvanceSeconds(100)
			timeNow := timeProvider.TimeNowUTC
			store.Sessions["session-id"] = session

			Convey("should be update accessed at time", func() {
				err := provider.Access(&session)
				So(err, ShouldBeNil)
				So(session.AccessedAt, ShouldEqual, timeNow)
			})
		})

		Convey("invalidating session", func() {
			store.Sessions["session-id"] = auth.Session{
				ID:                   "session-id",
				UserID:               "user-id",
				PrincipalID:          "principal-id",
				CreatedAt:            initialTime,
				AccessedAt:           initialTime,
				AccessToken:          "access-token",
				AccessTokenCreatedAt: initialTime,
			}

			Convey("should be successful", func() {
				err := provider.Invalidate("session-id")
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldBeEmpty)
			})

			Convey("should be successful for non-existant sessions", func() {
				err := provider.Invalidate("session-id-unknown")
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldNotBeEmpty)
			})
		})
	})
}
