package session

import (
	"math/rand"
	"net/http"
	"testing"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	Convey("Provider", t, func() {
		store := NewMockStore()

		timeProvider := &time.MockProvider{}
		initialTime := gotime.Date(2020, 1, 1, 0, 0, 0, 0, gotime.UTC)
		timeProvider.TimeNow = initialTime
		timeProvider.TimeNowUTC = initialTime

		req, _ := http.NewRequest("POST", "", nil)
		req.Header.Set("User-Agent", "SDK")
		req.Header.Set("X-Skygear-Extra-Info", "eyAiZGV2aWNlX25hbWUiOiAiRGV2aWNlIiB9")
		accessEvent := auth.AccessEvent{
			Timestamp: initialTime,
			UserAgent: "SDK",
			Extra: auth.AccessEventExtraInfo{
				"device_name": "Device",
			},
		}

		provider := &ProviderImpl{
			req:    req,
			store:  store,
			config: config.SessionConfiguration{},
			time:   timeProvider,
			rand:   rand.New(rand.NewSource(0)),
		}

		Convey("creating session", func() {
			Convey("should be successful", func() {
				session, token := provider.MakeSession(&authn.Attrs{
					UserID:      "user-id",
					PrincipalID: "principal-id",
				})
				err := provider.Create(session)

				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)
				So(session, ShouldResemble, &IDPSession{
					ID: session.ID,
					Attrs: authn.Attrs{
						UserID:      "user-id",
						PrincipalID: "principal-id",
					},
					AccessInfo: auth.AccessInfo{
						InitialAccess: accessEvent,
						LastAccess:    accessEvent,
					},
					CreatedAt: initialTime,
					TokenHash: session.TokenHash,
				})
			})

			Convey("should allow creating multiple sessions for same principal", func() {
				session1, _ := provider.MakeSession(&authn.Attrs{
					UserID:      "user-id",
					PrincipalID: "principal-id",
				})
				err := provider.Create(session1)
				So(err, ShouldBeNil)
				So(session1, ShouldResemble, &IDPSession{
					ID: session1.ID,
					Attrs: authn.Attrs{
						UserID:      "user-id",
						PrincipalID: "principal-id",
					},
					AccessInfo: auth.AccessInfo{
						InitialAccess: accessEvent,
						LastAccess:    accessEvent,
					},
					CreatedAt: initialTime,
					TokenHash: session1.TokenHash,
				})

				session2, _ := provider.MakeSession(&authn.Attrs{
					UserID:      "user-id",
					PrincipalID: "principal-id",
				})
				err = provider.Create(session2)
				So(err, ShouldBeNil)
				So(session2, ShouldResemble, &IDPSession{
					ID: session2.ID,
					Attrs: authn.Attrs{
						UserID:      "user-id",
						PrincipalID: "principal-id",
					},
					AccessInfo: auth.AccessInfo{
						InitialAccess: accessEvent,
						LastAccess:    accessEvent,
					},
					CreatedAt: initialTime,
					TokenHash: session2.TokenHash,
				})

				So(session1.ID, ShouldNotEqual, session2.ID)
				So(session1.TokenHash, ShouldNotEqual, session2.TokenHash)
			})
		})

		Convey("getting session", func() {
			fixtureSession := IDPSession{
				ID: "session-id",
				Attrs: authn.Attrs{
					UserID:      "user-id",
					PrincipalID: "principal-id",
				},
				CreatedAt: initialTime,
				TokenHash: "15be5b9c05673532b445d3295a86afd6b2615775e0233e9798cbe3c846a08d05",
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
			Convey("should reject if session is expired", func() {
				timeProvider.AdvanceSeconds(1000000)
				session, err := provider.GetByToken("session-id.token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
		})

		Convey("invalidating session", func() {
			store.Sessions["session-id"] = IDPSession{
				ID: "session-id",
				Attrs: authn.Attrs{
					UserID:      "user-id",
					PrincipalID: "principal-id",
				},
				CreatedAt: initialTime,
				TokenHash: "token-hash",
			}

			Convey("should be successful", func() {
				err := provider.Invalidate(&IDPSession{ID: "session-id"})
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldBeEmpty)
			})

			Convey("should be successful for non-existent sessions", func() {
				err := provider.Invalidate(&IDPSession{ID: "session-id-unknown"})
				So(err, ShouldBeNil)
				So(store.Sessions, ShouldNotBeEmpty)
			})
		})

		Convey("listing session", func() {
			makeSession := func(id string, userID string, timeOffset int) {
				store.Sessions[id] = IDPSession{
					ID: id,
					Attrs: authn.Attrs{
						UserID: userID,
					},
					CreatedAt: initialTime.Add(gotime.Duration(timeOffset) * gotime.Second),
				}
			}
			makeSession("a", "user-1", 100)
			makeSession("b", "user-1", 200)
			makeSession("c", "user-2", -10000)
			timeProvider.AdvanceSeconds(500)
			provider.config = config.SessionConfiguration{
				Lifetime: 1000,
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
}
