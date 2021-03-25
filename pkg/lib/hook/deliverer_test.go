package hook

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/jwk"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeliverer(t *testing.T) {
	Convey("Event Deliverer", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.HookConfig{
			SyncTimeout:      5,
			SyncTotalTimeout: 10,
		}
		key, err := jwk.New([]byte("aG9vay1zZWNyZXQ"))
		So(err, ShouldBeNil)
		set := jwk.NewSet()
		_ = set.Add(key)
		secret := &config.WebhookKeyMaterials{
			Set: set,
		}

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")

		httpClient := &http.Client{}
		gock.InterceptClient(httpClient)
		deliverer := Deliverer{
			Config:    cfg,
			Secret:    secret,
			Clock:     clock,
			SyncHTTP:  SyncHTTPClient{httpClient},
			AsyncHTTP: AsyncHTTPClient{httpClient},
		}

		defer gock.Off()

		Convey("determining whether the event will be delivered", func() {
			Convey("should return correct value for blocking events", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverBlockingEvent(blocking.PreSignup), ShouldBeTrue)
				So(deliverer.WillDeliverBlockingEvent(blocking.AdminAPICreateUser), ShouldBeFalse)
			})

			Convey("should return correct value for non-blocking events", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{
							string(nonblocking.UserCreatedAdminAPICreateUser),
							string(nonblocking.IdentityCreatedAdminAPIAddIdentity),
						},
						URL: "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.UserCreatedAdminAPICreateUser), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.IdentityCreatedAdminAPIAddIdentity), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.UserCreatedUserSignup), ShouldBeFalse)
			})

			Convey("should return true for all non-blocking events", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{"*"},
						URL:    "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.UserCreatedAdminAPICreateUser), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.IdentityCreatedAdminAPIAddIdentity), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.UserCreatedUserSignup), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(nonblocking.UserPromoted), ShouldBeTrue)
			})
		})

		Convey("delivering blocking events", func() {
			e := event.Event{
				ID:   "event-id",
				Type: blocking.PreSignup,
			}

			Convey("should be successful", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
					{
						Event: string(blocking.AdminAPICreateUser),
						URL:   "https://example.com/b",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					HeaderPresent(HeaderRequestBodySignature).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBlockingEvent(&e)

				So(err, ShouldBeNil)
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should disallow operation", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/b",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})

				gock.New("https://example.com").
					Post("/b").
					JSON(e).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": false,
						"reason":     "nope",
						"data": map[string]interface{}{
							"extra": 123,
						},
					})
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBlockingEvent(&e)

				So(err, ShouldBeError, "disallowed by web-hook event handler")
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should reject invalid status code", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(500)
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBlockingEvent(&e)

				So(err, ShouldBeError, "invalid status code")
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should time out long requests", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
					{
						Event: string(blocking.PreSignup),
						URL:   "https://example.com/a",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					Times(3).
					JSON(e).
					Reply(200).
					Map(func(resp *http.Response) *http.Response {
						clock.AdvanceSeconds(5)
						return resp
					}).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBlockingEvent(&e)

				So(err, ShouldBeError, "web-hook event delivery timed out")
				So(gock.IsDone(), ShouldBeTrue)
			})
		})

		Convey("delivering non-blocking events", func() {
			e := event.Event{
				ID:            "event-id",
				IsNonBlocking: true,
				Type:          nonblocking.UserCreatedUserSignup,
			}

			Convey("should be successful", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{string(nonblocking.UserCreatedUserSignup)},
						URL:    "https://example.com/a",
					},
					{
						Events: []string{string(nonblocking.UserCreatedAdminAPICreateUser)},
						URL:    "https://example.com/b",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(200).
					BodyString("test")
				defer func() { gock.Flush() }()

				err := deliverer.DeliverNonBlockingEvent(&e)

				So(err, ShouldBeNil)
				So(gock.IsDone(), ShouldBeTrue)
			})
		})
	})
}
