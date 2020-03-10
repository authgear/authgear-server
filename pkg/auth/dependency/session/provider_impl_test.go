package session

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	Convey("Provider", t, func() {
		store := NewMockStore()
		eventStore := NewMockEventStore()

		timeProvider := &time.MockProvider{}
		initialTime := gotime.Date(2020, 1, 1, 0, 0, 0, 0, gotime.UTC)
		timeProvider.TimeNow = initialTime
		timeProvider.TimeNowUTC = initialTime

		clientConfigs := []config.APIClientConfiguration{
			config.APIClientConfiguration{
				ID: "web-app",
			},
			config.APIClientConfiguration{
				ID: "mobile-app",
			},
		}

		req, _ := http.NewRequest("POST", "", nil)
		req.Header.Set("User-Agent", "SDK")
		req.Header.Set("X-Skygear-Extra-Info", "eyAiZGV2aWNlX25hbWUiOiAiRGV2aWNlIiB9")
		accessEvent := AccessEvent{
			Timestamp: initialTime,
			UserAgent: "SDK",
			Extra: AccessEventExtraInfo{
				"device_name": "Device",
			},
		}

		var provider Provider = &providerImpl{
			req:           req,
			store:         store,
			eventStore:    eventStore,
			clientConfigs: clientConfigs,
			time:          timeProvider,
			rand:          rand.New(rand.NewSource(0)),
		}

		Convey("creating session", func() {
			Convey("should be successful", func() {
				session, token := provider.MakeSession(&auth.AuthnSession{
					UserID:      "user-id",
					PrincipalID: "principal-id",
					ClientID:    "web-app",
				})
				err := provider.Create(session)

				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)
				So(session, ShouldResemble, &Session{
					ID:            session.ID,
					ClientID:      "web-app",
					UserID:        "user-id",
					PrincipalID:   "principal-id",
					InitialAccess: accessEvent,
					LastAccess:    accessEvent,
					CreatedAt:     initialTime,
					AccessedAt:    initialTime,
					TokenHash:     session.TokenHash,
				})
				So(eventStore.AccessEvents, ShouldResemble, []AccessEvent{accessEvent})
			})

			Convey("should allow creating multiple sessions for same principal", func() {
				session1, _ := provider.MakeSession(&auth.AuthnSession{
					UserID:      "user-id",
					PrincipalID: "principal-id",
					ClientID:    "web-app",
				})
				err := provider.Create(session1)
				So(err, ShouldBeNil)
				So(session1, ShouldResemble, &Session{
					ID:            session1.ID,
					ClientID:      "web-app",
					UserID:        "user-id",
					PrincipalID:   "principal-id",
					InitialAccess: accessEvent,
					LastAccess:    accessEvent,
					CreatedAt:     initialTime,
					AccessedAt:    initialTime,
					TokenHash:     session1.TokenHash,
				})

				session2, _ := provider.MakeSession(&auth.AuthnSession{
					UserID:      "user-id",
					PrincipalID: "principal-id",
					ClientID:    "web-app",
				})
				err = provider.Create(session2)
				So(err, ShouldBeNil)
				So(session2, ShouldResemble, &Session{
					ID:            session2.ID,
					ClientID:      "web-app",
					UserID:        "user-id",
					PrincipalID:   "principal-id",
					InitialAccess: accessEvent,
					LastAccess:    accessEvent,
					CreatedAt:     initialTime,
					AccessedAt:    initialTime,
					TokenHash:     session2.TokenHash,
				})

				So(session1.ID, ShouldNotEqual, session2.ID)
				So(session1.TokenHash, ShouldNotEqual, session2.TokenHash)
			})
		})

		Convey("getting session", func() {
			fixtureSession := Session{
				ID:          "session-id",
				ClientID:    "web-app",
				UserID:      "user-id",
				PrincipalID: "principal-id",
				CreatedAt:   initialTime,
				AccessedAt:  initialTime,
				TokenHash:   "15be5b9c05673532b445d3295a86afd6b2615775e0233e9798cbe3c846a08d05",
			}
			store.Sessions[fixtureSession.ID] = fixtureSession

			Convey("should be successful using session token", func() {
				session, err := provider.GetByToken("session-id.token")
				So(err, ShouldBeNil)
				So(session, ShouldResemble, &fixtureSession)
			})

			Convey("should reject non-existent session", func() {
				session, err := provider.GetByToken("session-id-unknown.token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})

			Convey("should reject incorrect token", func() {
				session, err := provider.GetByToken("session-id.incorrect-token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)

				session, err = provider.GetByToken("invalid-token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
			Convey("should reject if client does not exists", func(c C) {
				for i := range clientConfigs {
					if clientConfigs[i].ID == "web-app" {
						clientConfigs[i].ID = "node-app"
					}
				}
				session, err := provider.GetByToken("session-id.token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
			Convey("should reject if session is expired", func() {
				timeProvider.AdvanceSeconds(1000000)
				session, err := provider.GetByToken("session-id.token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
		})

		Convey("accessing session", func() {
			session := Session{
				ID:          "session-id",
				UserID:      "user-id",
				PrincipalID: "principal-id",
				CreatedAt:   initialTime,
				AccessedAt:  initialTime,
				TokenHash:   "token-hash",
				ClientID:    "web-app",
			}
			timeProvider.AdvanceSeconds(100)
			timeNow := timeProvider.TimeNowUTC
			accessEvent.Timestamp = timeNow
			store.Sessions["session-id"] = session

			Convey("should be update accessed at time", func() {
				err := provider.Access(&session)
				So(err, ShouldBeNil)
				So(session.AccessedAt, ShouldEqual, timeNow)
			})
			Convey("should be create access event", func() {
				err := provider.Access(&session)
				So(err, ShouldBeNil)
				So(session.LastAccess, ShouldResemble, accessEvent)
				So(eventStore.AccessEvents, ShouldResemble, []AccessEvent{accessEvent})
			})
		})

		Convey("invalidating session", func() {
			store.Sessions["session-id"] = Session{
				ID:          "session-id",
				UserID:      "user-id",
				PrincipalID: "principal-id",
				CreatedAt:   initialTime,
				AccessedAt:  initialTime,
				TokenHash:   "token-hash",
			}

			Convey("should be successful", func() {
				err := provider.Invalidate(&Session{ID: "session-id"})
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldBeEmpty)
			})

			Convey("should be successful for non-existent sessions", func() {
				err := provider.Invalidate(&Session{ID: "session-id-unknown"})
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldNotBeEmpty)
			})
		})

		Convey("listing session", func() {
			makeSession := func(id string, userID string, clientID string, timeOffset int) {
				store.Sessions[id] = Session{
					ID:         id,
					UserID:     userID,
					ClientID:   clientID,
					CreatedAt:  initialTime.Add(gotime.Duration(timeOffset) * gotime.Second),
					AccessedAt: initialTime.Add(gotime.Duration(timeOffset) * gotime.Second),
				}
			}
			makeSession("a", "user-1", "web-app", 100)
			makeSession("b", "user-1", "mobile-app", 200)
			makeSession("c", "user-2", "web-app", -10000)
			makeSession("d", "user-2", "disabled-app", 400)
			timeProvider.AdvanceSeconds(500)
			for i := range clientConfigs {
				clientConfigs[i] = config.APIClientConfiguration{
					ID:                   clientConfigs[i].ID,
					RefreshTokenLifetime: 1000,
					RefreshTokenDisabled: true,
				}
			}

			list := func(userID string) (ids []string, err error) {
				sessions, err := provider.List(userID)
				for _, session := range sessions {
					ids = append(ids, session.ID)
				}
				return
			}

			Convey("should be correctly filtered", func() {
				ids, err := list("user-1")
				So(err, ShouldBeNil)
				So(ids, ShouldResemble, []string{"a", "b"})

				ids, err = list("user-2")
				So(err, ShouldBeNil)
				So(ids, ShouldHaveLength, 0)
			})
		})
	})
	Convey("newAccessEvent", t, func() {
		now := gotime.Date(2020, 1, 1, 0, 0, 0, 0, gotime.UTC)
		Convey("should record current timestamp", func() {
			req, _ := http.NewRequest("POST", "", nil)

			event := newAccessEvent(now, req)
			So(event.Timestamp, ShouldResemble, now)
		})
		Convey("should populate connection info", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("X-Forwarded-For", "13.225.103.28, 216.58.197.110")
			req.Header.Set("X-Real-IP", "216.58.197.110")
			req.Header.Set("Forwarded", "for=216.58.197.110;proto=http;by=192.168.1.11")

			event := newAccessEvent(now, req)
			So(event.Remote, ShouldResemble, AccessEventConnInfo{
				RemoteAddr:    "192.168.1.11:31035",
				XForwardedFor: "13.225.103.28, 216.58.197.110",
				XRealIP:       "216.58.197.110",
				Forwarded:     "for=216.58.197.110;proto=http;by=192.168.1.11",
			})
		})
		Convey("should populate user agent", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("User-Agent", "SDK")

			event := newAccessEvent(now, req)
			So(event.UserAgent, ShouldEqual, "SDK")
		})
		Convey("should populate extra info", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.Header.Set("X-Skygear-Extra-Info", "eyAiZGV2aWNlX25hbWUiOiAiRGV2aWNlIiB9")

			event := newAccessEvent(now, req)
			So(event.Extra, ShouldResemble, AccessEventExtraInfo{
				"device_name": "Device",
			})
		})
		Convey("should not populate extra info if too large", func() {
			extra := "{ "
			for i := 0; i < 1000; i++ {
				if i != 0 {
					extra += ", "
				}
				extra += fmt.Sprintf(`"info_%d": %d`, i, i)
			}
			extra += " }"
			extra = base64.StdEncoding.EncodeToString([]byte(extra))

			req, _ := http.NewRequest("POST", "", nil)
			req.Header.Set("X-Skygear-Extra-Info", extra)

			event := newAccessEvent(now, req)
			So(event.Extra, ShouldResemble, AccessEventExtraInfo{})
		})
	})
}
