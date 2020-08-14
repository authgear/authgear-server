package idpsession

import (
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

type mockAccessEventProvider struct{}

func (*mockAccessEventProvider) InitStream(sessionID string, event *access.Event) error {
	return nil
}

func TestProvider(t *testing.T) {
	Convey("Provider", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := NewMockStore(ctrl)

		clock := clock.NewMockClockAt("2020-01-01T00:00:00Z")
		initialTime := clock.Time

		req, _ := http.NewRequest("POST", "", nil)
		req.Header.Set("User-Agent", "SDK")
		req.Header.Set("X-Authgear-Extra-Info", "eyAiZGV2aWNlX25hbWUiOiAiRGV2aWNlIiB9")
		accessEvent := access.Event{
			Timestamp: initialTime,
			UserAgent: "SDK",
			Extra: access.EventExtraInfo{
				"device_name": "Device",
			},
		}

		provider := &Provider{
			Request:      req,
			Store:        store,
			AccessEvents: &mockAccessEventProvider{},
			ServerConfig: &config.ServerConfig{},
			Config:       &config.SessionConfig{},
			Clock:        clock,
			Random:       rand.New(rand.NewSource(0)),
		}

		Convey("creating session", func() {
			Convey("should be successful", func() {
				store.EXPECT().Create(gomock.Any(), initialTime).Return(nil)

				s, token := provider.MakeSession(&session.Attrs{
					UserID: "user-id",
				})
				err := provider.Create(s)

				So(err, ShouldBeNil)
				So(token, ShouldNotBeEmpty)
				So(s, ShouldResemble, &IDPSession{
					ID: s.ID,
					Attrs: session.Attrs{
						UserID: "user-id",
					},
					AccessInfo: access.Info{
						InitialAccess: accessEvent,
						LastAccess:    accessEvent,
					},
					CreatedAt: initialTime,
					TokenHash: s.TokenHash,
				})
			})
		})

		Convey("getting session", func() {
			fixtureSession := IDPSession{
				ID: "session-id",
				Attrs: session.Attrs{
					UserID: "user-id",
				},
				CreatedAt: initialTime,
				TokenHash: "15be5b9c05673532b445d3295a86afd6b2615775e0233e9798cbe3c846a08d05",
			}
			store.EXPECT().Get(gomock.Any()).DoAndReturn(func(id string) (*IDPSession, error) {
				if id == fixtureSession.ID {
					return &fixtureSession, nil
				}
				return nil, ErrSessionNotFound
			})

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
				clock.AdvanceSeconds(1000000)
				session, err := provider.GetByToken("session-id.token")
				So(err, ShouldBeError, ErrSessionNotFound)
				So(session, ShouldBeNil)
			})
		})
	})
}
