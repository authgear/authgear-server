package hook

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDispatchEvent(t *testing.T) {
	Convey("Hook Provider", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
		store := NewMockStore(ctrl)
		deliverer := NewMockDeliverer(ctrl)
		users := NewMockUserProvider(ctrl)
		db := NewMockDatabaseHandle(ctrl)
		ctx := context.Background()

		fallbackLanguage := "en"
		supportedLanguages := []string{"en"}
		provider := &Provider{
			Context:   ctx,
			Logger:    Logger{log.Null},
			Database:  db,
			Clock:     clock,
			Users:     users,
			Store:     store,
			Deliverer: deliverer,
			Localization: &config.LocalizationConfig{
				FallbackLanguage:   &fallbackLanguage,
				SupportedLanguages: supportedLanguages,
			},
		}

		var seq int64 = 0
		store.EXPECT().NextSequenceNumber().AnyTimes().
			DoAndReturn(func() (int64, error) {
				seq++
				return seq, nil
			})

		db.EXPECT().UseHook(provider).AnyTimes()

		Convey("dispatching blocking events", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &blocking.PreSignupBlockingEvent{
				User: user,
			}

			Convey("should be successful", func() {
				deliverer.EXPECT().WillDeliverBlockingEvent(blocking.PreSignup).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    blocking.PreSignup,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(nil)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeEmpty)
			})

			Convey("should not generate before events that would not be delivered", func() {
				deliverer.EXPECT().WillDeliverBlockingEvent(blocking.PreSignup).Return(false)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeEmpty)
			})

			Convey("should include auth info", func() {
				userID := "user-id"
				provider.Context = session.WithSession(
					context.Background(),
					&idpsession.IDPSession{
						ID: "user-id-principal-id",
						Attrs: session.Attrs{
							UserID: userID,
						},
					},
				)

				deliverer.EXPECT().WillDeliverBlockingEvent(blocking.PreSignup).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    blocking.PreSignup,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           &userID,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(nil)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
			})

			Convey("should return delivery error", func() {
				deliverer.EXPECT().WillDeliverBlockingEvent(blocking.PreSignup).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(gomock.Any()).
					Return(fmt.Errorf("failed to deliver"))

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeError, "failed to dispatch event: failed to deliver")
			})
		})

		Convey("dispatching non-blocking events", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &nonblocking.UserCreatedUserSignupEvent{
				User: user,
			}

			Convey("should be successful", func() {
				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})
		})

		Convey("when transaction is about to commit", func() {
			Convey("should generate & persist events", func() {
				payload := &nonblocking.UserCreatedUserSignupEvent{
					User: model.User{
						Meta: model.Meta{ID: "user-id"},
					},
				}
				provider.persistentEventPayloads = []event.Payload{
					payload,
				}
				deliverer.EXPECT().WillDeliverNonBlockingEvent(payload.NonBlockingEventType()).Return(true)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:      "0000000000000001",
						Type:    payload.NonBlockingEventType(),
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
						IsNonBlocking: true,
					},
				})

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeNil)
			})

			Convey("should not generate events that would not be delivered", func() {
				provider.persistentEventPayloads = []event.Payload{
					&nonblocking.UserCreatedUserSignupEvent{
						User: model.User{
							Meta: model.Meta{ID: "user-id"},
						},
					},
				}
				deliverer.EXPECT().WillDeliverNonBlockingEvent(nonblocking.UserCreatedUserSignup).Return(false)
				store.EXPECT().AddEvents([]*event.Event{})

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeNil)
			})
		})

		Convey("should skip db hook if there is error during dispatch error", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &blocking.PreSignupBlockingEvent{
				User: user,
			}
			payload2 := &blocking.PreSignupBlockingEvent{
				User: user,
			}
			nonBlockingPayload := &nonblocking.UserCreatedUserSignupEvent{
				User: user,
			}
			webhookErr := WebHookDisallowed.New("")

			Convey("should call db hook function when dispatch event success", func() {
				deliverer.EXPECT().WillDeliverBlockingEvent(payload.BlockingEventType()).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    payload.BlockingEventType(),
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(nil)

				// Calling provider.WillCommitTx will trigger persistent events
				deliverer.EXPECT().WillDeliverNonBlockingEvent(nonBlockingPayload.NonBlockingEventType()).Return(true)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:      "0000000000000002",
						Type:    nonBlockingPayload.NonBlockingEventType(),
						Seq:     2,
						Payload: nonBlockingPayload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
						IsNonBlocking: true,
					},
				})

				err := provider.DispatchEvent(payload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.DispatchEvent(nonBlockingPayload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.WillCommitTx()
				So(err, ShouldBeNil)
			})

			Convey("should not add db hook", func() {
				deliverer.EXPECT().WillDeliverBlockingEvent(payload.BlockingEventType()).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    payload.BlockingEventType(),
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(webhookErr)

				err := provider.DispatchEvent(payload)
				So(err, ShouldBeError, webhookErr)
				So(provider.dbHooked, ShouldBeTrue)

				// Calling provider.WillCommitTx will not trigger persistent events
				err = provider.WillCommitTx()
				So(err, ShouldBeNil)
			})

			Convey("should not generate events that would not be delivered", func() {
				// first event
				deliverer.EXPECT().WillDeliverBlockingEvent(payload.BlockingEventType()).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    payload.BlockingEventType(),
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(nil)

				// second event
				deliverer.EXPECT().WillDeliverBlockingEvent(payload2.BlockingEventType()).Return(true)
				deliverer.EXPECT().DeliverBlockingEvent(
					&event.Event{
						ID:      "0000000000000002",
						Type:    payload2.BlockingEventType(),
						Seq:     2,
						Payload: payload2,
						Context: event.Context{
							Timestamp:        1136214245,
							UserID:           nil,
							ResolvedLanguage: fallbackLanguage,
						},
					},
				).Return(webhookErr)

				err := provider.DispatchEvent(payload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.DispatchEvent(payload2)
				So(err, ShouldBeError, webhookErr)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.DispatchEvent(nonBlockingPayload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				// Calling provider.WillCommitTx will not trigger persistent events
				err = provider.WillCommitTx()
				So(err, ShouldBeNil)
			})
		})
	})
}
