package hook

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/jwk"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
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
		stdAttrsService := NewMockStdAttrsServiceNoEvent(ctrl)

		deliverer := Deliverer{
			Config:                 cfg,
			Secret:                 secret,
			Clock:                  clock,
			SyncHTTP:               SyncHTTPClient{httpClient},
			AsyncHTTP:              AsyncHTTPClient{httpClient},
			StdAttrsServiceNoEvent: stdAttrsService,
		}

		defer gock.Off()

		Convey("determining whether the event will be delivered", func() {
			Convey("should return correct value for blocking events", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverBlockingEvent(MockBlockingEventType1), ShouldBeTrue)
				So(deliverer.WillDeliverBlockingEvent(MockBlockingEventType2), ShouldBeFalse)
			})

			Convey("should return correct value for non-blocking events", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{
							string(MockNonBlockingEventType1),
							string(MockNonBlockingEventType2),
						},
						URL: "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType1), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType2), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType3), ShouldBeFalse)
			})

			Convey("should return true for all non-blocking events", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{"*"},
						URL:    "https://example.com/a",
					},
				}

				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType1), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType2), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType3), ShouldBeTrue)
				So(deliverer.WillDeliverNonBlockingEvent(MockNonBlockingEventType4), ShouldBeTrue)
			})
		})

		Convey("delivering blocking events", func() {
			e := event.Event{
				ID:   "event-id",
				Type: MockBlockingEventType1,
			}

			Convey("should be successful", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
					{
						Event: string(MockBlockingEventType2),
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

			Convey("should apply mutations along the chain", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/do-not-mutate",
					},
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/mutate-something",
					},
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/see-mutated-thing",
					},
				}

				originalEvent := &event.Event{
					ID:      "event-id",
					Type:    MockBlockingEventType1,
					Payload: &MockBlockingEvent1{},
				}

				mutatedEvent := &event.Event{
					ID:   "event-id",
					Type: MockBlockingEventType1,
					Payload: &MockBlockingEvent1{
						MockUserEventBase: MockUserEventBase{
							User: model.User{
								StandardAttributes: map[string]interface{}{
									"name": "John Doe",
								},
							},
						},
					},
				}

				gock.New("https://example.com").
					Post("/do-not-mutate").
					JSON(originalEvent).
					HeaderPresent(HeaderRequestBodySignature).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})

				gock.New("https://example.com").
					Post("/mutate-something").
					JSON(originalEvent).
					HeaderPresent(HeaderRequestBodySignature).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
						"mutations": map[string]interface{}{
							"user": map[string]interface{}{
								"standard_attributes": map[string]interface{}{
									"name": "John Doe",
								},
							},
						},
					})

				gock.New("https://example.com").
					Post("/see-mutated-thing").
					JSON(mutatedEvent).
					HeaderPresent(HeaderRequestBodySignature).
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})

				defer func() { gock.Flush() }()

				stdAttrsService.EXPECT().UpdateStandardAttributes(
					config.RolePortalUI,
					gomock.Any(),
					map[string]interface{}{
						"name": "John Doe",
					},
				).Times(1).Return(nil)

				err := deliverer.DeliverBlockingEvent(originalEvent)

				So(err, ShouldBeNil)
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should disallow operation", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
					{
						Event: string(MockBlockingEventType1),
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
					})
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBlockingEvent(&e)

				So(err, ShouldBeError, "disallowed by web-hook event handler")
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should reject invalid status code", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
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
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
					{
						Event: string(MockBlockingEventType1),
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

				So(err, ShouldBeError, "webhook delivery timeout")
				So(gock.IsDone(), ShouldBeTrue)
			})
		})

		Convey("delivering non-blocking events", func() {
			e := event.Event{
				ID:            "event-id",
				IsNonBlocking: true,
				Type:          MockNonBlockingEventType3,
			}

			Convey("should be successful", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{string(MockNonBlockingEventType3)},
						URL:    "https://example.com/a",
					},
					{
						Events: []string{string(MockNonBlockingEventType1)},
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
