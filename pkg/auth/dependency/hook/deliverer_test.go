package hook

import (
	"fmt"
	gohttp "net/http"
	"testing"
	gotime "time"

	"github.com/golang/mock/gomock"
	"github.com/h2non/gock"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeliverer(t *testing.T) {
	Convey("Event Deliverer", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		hookAppConfig := &config.HookAppConfiguration{
			Secret: "hook-secret",
		}
		hookTenantConfig := &config.HookTenantConfiguration{
			SyncHookTimeout:      5,
			SyncHookTotalTimeout: 10,
		}

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")

		m := NewMockMutator(ctrl)
		mf := NewMockMutatorFactory(ctrl)

		httpClient := gohttp.Client{}
		gock.InterceptClient(&httpClient)
		deliverer := Deliverer{
			HookAppConfig:    hookAppConfig,
			HookTenantConfig: hookTenantConfig,
			Clock:            clock,
			MutatorFactory:   mf,
			HTTPClient:       httpClient,
		}

		defer gock.Off()

		Convey("determining whether the event will be delivered", func() {
			Convey("should return correct value", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.UserSync),
						URL:   "https://example.com/b",
					},
				}

				So(deliverer.WillDeliver(event.BeforeSessionCreate), ShouldBeTrue)
				So(deliverer.WillDeliver(event.UserSync), ShouldBeTrue)
				So(deliverer.WillDeliver(event.AfterSessionCreate), ShouldBeFalse)
			})
		})

		Convey("delivering before events", func() {
			e := event.Event{
				ID:   "event-id",
				Type: event.BeforeSessionCreate,
			}

			var user model.User
			mf.EXPECT().New(&e, &user).Return(m)

			Convey("should be successful", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionDelete),
						URL:   "https://example.com/b",
					},
				}

				user = model.User{
					ID: "user-id",
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

				m.EXPECT().Apply().Return(nil)

				err := deliverer.DeliverBeforeEvent(&e, &user)

				So(err, ShouldBeNil)
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should disallow operation", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/b",
					},
				}

				user = model.User{
					ID: "user-id",
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

				err := deliverer.DeliverBeforeEvent(&e, &user)

				So(err, ShouldBeError, "disallowed by web-hook event handler")
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should apply mutations", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/b",
					},
				}

				user = model.User{
					ID: "user-id",
					Metadata: map[string]interface{}{
						"test": 123,
					},
				}

				e = event.Event{
					ID:   "event-id",
					Type: event.BeforeSessionCreate,
					Payload: event.SessionCreateEvent{
						User: model.User{
							ID: "user-id",
							Metadata: map[string]interface{}{
								"test": 123,
							},
						},
						Identity: model.Identity{},
					},
				}

				gock.New("https://example.com").
					Post("/a").
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
						"mutations": map[string]interface{}{
							"metadata": map[string]interface{}{
								"test1": 123,
							},
						},
					})

				gock.New("https://example.com").
					Post("/b").
					Reply(200).
					JSON(map[string]interface{}{
						"is_allowed": true,
						"mutations": map[string]interface{}{
							"metadata": map[string]interface{}{
								"test2": true,
							},
						},
					})
				defer func() { gock.Flush() }()

				Convey("successful", func() {
					m.EXPECT().Add(gomock.Eq(event.Mutations{
						Metadata: &map[string]interface{}{
							"test1": float64(123),
						},
					})).Return(nil)
					m.EXPECT().Add(gomock.Eq(event.Mutations{
						Metadata: &map[string]interface{}{
							"test2": true,
						},
					})).Return(nil)
					m.EXPECT().Apply().Return(nil)

					err := deliverer.DeliverBeforeEvent(&e, &user)

					So(err, ShouldBeNil)
					So(gock.IsDone(), ShouldBeTrue)
				})

				Convey("failed apply", func() {
					m.EXPECT().Add(event.Mutations{
						Metadata: &map[string]interface{}{
							"test1": float64(123),
						},
					}).Return(nil)
					m.EXPECT().Add(event.Mutations{
						Metadata: &map[string]interface{}{
							"test2": true,
						},
					}).Return(nil)
					m.EXPECT().Apply().Return(fmt.Errorf("cannot apply mutations"))

					err := deliverer.DeliverBeforeEvent(&e, &user)

					So(err, ShouldBeError, "web-hook mutation failed: cannot apply mutations")
					So(gock.IsDone(), ShouldBeTrue)
				})

				Convey("failed add", func() {
					m.EXPECT().Add(event.Mutations{
						Metadata: &map[string]interface{}{
							"test1": float64(123),
						},
					}).Return(fmt.Errorf("cannot add mutations"))
					err := deliverer.DeliverBeforeEvent(&e, &user)

					So(err, ShouldBeError, "web-hook mutation failed: cannot add mutations")
					So(gock.IsDone(), ShouldBeFalse)
				})
			})

			Convey("should reject invalid status code", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
				}

				user = model.User{
					ID: "user-id",
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(500)
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBeforeEvent(&e, &user)

				So(err, ShouldBeError, "invalid status code")
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should time out long requests", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
				}

				user = model.User{
					ID: "user-id",
				}

				gock.New("https://example.com").
					Post("/a").
					Times(3).
					JSON(e).
					Reply(200).
					Map(func(resp *gohttp.Response) *gohttp.Response {
						clock.AdvanceSeconds(5)
						return resp
					}).
					JSON(map[string]interface{}{
						"is_allowed": true,
					})
				defer func() { gock.Flush() }()

				err := deliverer.DeliverBeforeEvent(&e, &user)

				So(err, ShouldBeError, "web-hook event delivery timed out")
				So(gock.IsDone(), ShouldBeTrue)
			})
		})

		Convey("delivering non-before events", func() {
			e := event.Event{
				ID:   "event-id",
				Type: event.UserSync,
			}

			Convey("should be successful", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.UserSync),
						URL:   "https://example.com/a",
					},
					config.Hook{
						Event: string(event.AfterIdentityCreate),
						URL:   "https://example.com/b",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(200).
					BodyString("test")
				defer func() { gock.Flush() }()

				err := deliverer.DeliverNonBeforeEvent(&e, 5*gotime.Second)

				So(err, ShouldBeNil)
				So(gock.IsDone(), ShouldBeTrue)
			})

			Convey("should reject invalid status code", func() {
				deliverer.Hooks = &[]config.Hook{
					config.Hook{
						Event: string(event.UserSync),
						URL:   "https://example.com/a",
					},
				}

				gock.New("https://example.com").
					Post("/a").
					JSON(e).
					Reply(500)
				defer func() { gock.Flush() }()

				err := deliverer.DeliverNonBeforeEvent(&e, 5*gotime.Second)

				So(err, ShouldBeError, "invalid status code")
				So(gock.IsDone(), ShouldBeTrue)
			})
		})
	})
}
