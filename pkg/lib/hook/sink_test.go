package hook

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSink(t *testing.T) {
	mustURL := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	Convey("Sink", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := &config.HookConfig{
			SyncTimeout:      5,
			SyncTotalTimeout: 10,
		}

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")

		stdAttrsService := NewMockStandardAttributesServiceNoEvent(ctrl)
		customAttrsService := NewMockCustomAttributesServiceNoEvent(ctrl)
		webhook := NewMockEventWebHook(ctrl)
		denohook := NewMockEventDenoHook(ctrl)

		s := Sink{
			Config:             cfg,
			Clock:              clock,
			EventWebHook:       webhook,
			EventDenoHook:      denohook,
			StandardAttributes: stdAttrsService,
			CustomAttributes:   customAttrsService,
		}

		Convey("determining whether the event will be delivered", func() {
			Convey("should return correct value for blocking events", func() {
				cfg.BlockingHandlers = []config.BlockingHandlersConfig{
					{
						Event: string(MockBlockingEventType1),
						URL:   "https://example.com/a",
					},
				}

				So(s.WillDeliverBlockingEvent(MockBlockingEventType1), ShouldBeTrue)
				So(s.WillDeliverBlockingEvent(MockBlockingEventType2), ShouldBeFalse)
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

				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType1), ShouldBeTrue)
				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType2), ShouldBeTrue)
				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType3), ShouldBeFalse)
			})

			Convey("should return true for all non-blocking events", func() {
				cfg.NonBlockingHandlers = []config.NonBlockingHandlersConfig{
					{
						Events: []string{"*"},
						URL:    "https://example.com/a",
					},
				}

				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType1), ShouldBeTrue)
				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType2), ShouldBeTrue)
				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType3), ShouldBeTrue)
				So(s.WillDeliverNonBlockingEvent(MockNonBlockingEventType4), ShouldBeTrue)
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

				ctx := context.Background()

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[0].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(gomock.Any(), mustURL(cfg.BlockingHandlers[0].URL), &e).Times(1).Return(&event.HookResponse{
					IsAllowed: true,
				}, nil)

				err := s.DeliverBlockingEvent(ctx, &e)

				So(err, ShouldBeNil)
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
								CustomAttributes: map[string]interface{}{
									"a": "a",
								},
							},
						},
					},
				}

				ctx := context.Background()

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[0].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[0].URL),
					originalEvent,
				).Times(1).Return(
					&event.HookResponse{
						IsAllowed: true,
					},
					nil,
				)

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[1].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[1].URL),
					originalEvent,
				).Times(1).Return(
					&event.HookResponse{
						IsAllowed: true,
						Mutations: event.Mutations{
							User: event.UserMutations{
								StandardAttributes: map[string]interface{}{
									"name": "John Doe",
								},
								CustomAttributes: map[string]interface{}{
									"a": "a",
								},
							},
						},
					},
					nil,
				)

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[2].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[2].URL),
					mutatedEvent,
				).Times(1).Return(
					&event.HookResponse{
						IsAllowed: true,
					},
					nil,
				)

				stdAttrsService.EXPECT().UpdateStandardAttributes(
					ctx,
					accesscontrol.RoleGreatest,
					gomock.Any(),
					map[string]interface{}{
						"name": "John Doe",
					},
				).Times(1).Return(nil)

				customAttrsService.EXPECT().UpdateAllCustomAttributes(
					ctx,
					accesscontrol.RoleGreatest,
					gomock.Any(),
					map[string]interface{}{
						"a": "a",
					},
				).Times(1).Return(nil)

				err := s.DeliverBlockingEvent(ctx, originalEvent)

				So(err, ShouldBeNil)
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

				ctx := context.Background()

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[0].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[0].URL),
					&e,
				).Times(1).Return(
					&event.HookResponse{
						IsAllowed: true,
					},
					nil,
				)

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[1].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[1].URL),
					&e,
				).Times(1).Return(
					&event.HookResponse{
						IsAllowed: false,
						Reason:    "nope",
					},
					nil,
				)

				err := s.DeliverBlockingEvent(ctx, &e)

				So(err, ShouldBeError, "disallowed by web-hook event handler")
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

				ctx := context.Background()

				webhook.EXPECT().SupportURL(mustURL(cfg.BlockingHandlers[0].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverBlockingEvent(
					gomock.Any(),
					mustURL(cfg.BlockingHandlers[0].URL),
					&e,
				).AnyTimes().DoAndReturn(func(_ context.Context, _ *url.URL, _ *event.Event) (*event.HookResponse, error) {
					clock.AdvanceSeconds(5)
					return &event.HookResponse{
						IsAllowed: true,
					}, nil
				})

				err := s.DeliverBlockingEvent(ctx, &e)

				So(err, ShouldBeError, "webhook delivery timeout")
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

				webhook.EXPECT().SupportURL(mustURL(cfg.NonBlockingHandlers[0].URL)).AnyTimes().Return(true)
				webhook.EXPECT().DeliverNonBlockingEvent(
					gomock.Any(),
					mustURL(cfg.NonBlockingHandlers[0].URL),
					&e,
				).Times(1).Return(nil)

				ctx := context.Background()
				err := s.DeliverNonBlockingEvent(ctx, &e)

				So(err, ShouldBeNil)
			})
		})
	})
}
