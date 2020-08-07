package hook

import (
	"fmt"
	"net/http"
	"testing"
	gotime "time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/jwk"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"

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
		secret := &config.WebhookKeyMaterials{
			Set: jwk.Set{
				Keys: []jwk.Key{key},
			},
		}

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")

		m := NewMockMutator(ctrl)
		mf := NewMockMutatorFactory(ctrl)

		httpClient := &http.Client{}
		gock.InterceptClient(httpClient)
		deliverer := Deliverer{
			Config:         cfg,
			Secret:         secret,
			Clock:          clock,
			MutatorFactory: mf,
			SyncHTTP:       SyncHTTPClient{httpClient},
			AsyncHTTP:      AsyncHTTPClient{httpClient},
		}

		defer gock.Off()

		Convey("determining whether the event will be delivered", func() {
			Convey("should return correct value", func() {
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
						Event: string(event.BeforeSessionCreate),
						URL:   "https://example.com/a",
					},
					{
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
					Map(func(resp *http.Response) *http.Response {
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
						Event: string(event.UserSync),
						URL:   "https://example.com/a",
					},
					{
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
				cfg.Handlers = []config.HookHandlerConfig{
					{
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
