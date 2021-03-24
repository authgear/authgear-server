package hook

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
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

		provider := &Provider{
			Context:   ctx,
			Logger:    Logger{log.Null},
			Database:  db,
			Clock:     clock,
			Users:     users,
			Store:     store,
			Deliverer: deliverer,
		}

		var seq int64 = 0
		store.EXPECT().NextSequenceNumber().AnyTimes().
			DoAndReturn(func() (int64, error) {
				seq++
				return seq, nil
			})

		db.EXPECT().UseHook(provider).AnyTimes()

		Convey("dispatching operation events", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &event.SessionCreateEvent{
				Reason: "login",
				User:   user,
			}

			Convey("should be successful", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				).Return(nil)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})

			Convey("should not generate before events that would not be delivered", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(false)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})

			Convey("should use mutated payload", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				).DoAndReturn(func(ev *event.Event) error {
					payload := ev.Payload.(*event.SessionCreateEvent)
					payload.Reason = "signup"
					ev.Payload = payload
					return nil
				})

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					&event.SessionCreateEvent{
						Reason: "signup",
						User:   user,
					},
				})
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

				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    &userID,
						},
					},
				).Return(nil)

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeNil)
			})

			Convey("should return delivery error", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(gomock.Any()).
					Return(fmt.Errorf("failed to deliver"))

				err := provider.DispatchEvent(payload)

				So(err, ShouldBeError, "failed to dispatch event")
			})
		})

		Convey("dispatching notification events", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &event.UserSyncEvent{
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
				provider.persistentEventPayloads = []event.Payload{
					&event.SessionCreateEvent{
						User: model.User{
							Meta: model.Meta{ID: "user-id"},
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					Meta: model.Meta{ID: "user-id"},
				}, nil)
				deliverer.EXPECT().WillDeliver(event.UserSync).Return(true)
				deliverer.EXPECT().WillDeliver(event.AfterSessionCreate).Return(true)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:   "0000000000000001",
						Type: event.AfterSessionCreate,
						Seq:  1,
						Payload: &event.SessionCreateEvent{
							User: model.User{
								Meta: model.Meta{ID: "user-id"},
							},
						},
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
					{
						ID:   "0000000000000002",
						Type: event.UserSync,
						Seq:  2,
						Payload: &event.UserSyncEvent{
							User: model.User{
								Meta: model.Meta{ID: "user-id"},
							},
						},
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				})

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeNil)
			})

			Convey("should not generate events that would not be delivered", func() {
				provider.persistentEventPayloads = []event.Payload{
					&event.SessionCreateEvent{
						User: model.User{
							Meta: model.Meta{ID: "user-id"},
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					Meta: model.Meta{ID: "user-id"},
				}, nil)
				deliverer.EXPECT().WillDeliver(event.UserSync).Return(true)
				deliverer.EXPECT().WillDeliver(event.AfterSessionCreate).Return(false)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:   "0000000000000001",
						Type: event.UserSync,
						Seq:  1,
						Payload: &event.UserSyncEvent{
							User: model.User{
								Meta: model.Meta{ID: "user-id"},
							},
						},
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				})

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldBeNil)
			})
		})

		Convey("should skip db hook if there is error during dispatch error", func() {
			user := model.User{
				Meta: model.Meta{ID: "user-id"},
			}
			payload := &event.SessionCreateEvent{
				Reason: "login",
				User:   user,
			}
			payload2 := &event.UserCreateEvent{
				User: user,
			}
			webhookErr := WebHookDisallowed.New("")

			Convey("should call db hook function when dispatch event success", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				).Return(nil)

				// Calling provider.WillCommitTx will trigger persistent events
				users.EXPECT().Get("user-id").Return(&model.User{
					Meta: model.Meta{ID: "user-id"},
				}, nil)
				deliverer.EXPECT().WillDeliver(event.UserSync).Return(true)
				deliverer.EXPECT().WillDeliver(event.AfterSessionCreate).Return(true)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:      "0000000000000002",
						Type:    event.AfterSessionCreate,
						Seq:     2,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
					{
						ID:   "0000000000000003",
						Type: event.UserSync,
						Seq:  3,
						Payload: &event.UserSyncEvent{
							User: user,
						},
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				})

				err := provider.DispatchEvent(payload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.WillCommitTx()
				So(err, ShouldBeNil)
			})

			Convey("should not add db hook", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
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
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000001",
						Type:    event.BeforeSessionCreate,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				).Return(nil)

				// second event
				deliverer.EXPECT().WillDeliver(event.BeforeUserCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(
					&event.Event{
						ID:      "0000000000000002",
						Type:    event.BeforeUserCreate,
						Seq:     2,
						Payload: payload2,
						Context: event.Context{
							Timestamp: 1136214245,
							UserID:    nil,
						},
					},
				).Return(webhookErr)

				err := provider.DispatchEvent(payload)
				So(err, ShouldBeNil)
				So(provider.dbHooked, ShouldBeTrue)

				err = provider.DispatchEvent(payload2)
				So(err, ShouldBeError, webhookErr)
				So(provider.dbHooked, ShouldBeTrue)

				// Calling provider.WillCommitTx will not trigger persistent events
				err = provider.WillCommitTx()
				So(err, ShouldBeNil)
			})
		})
	})
}
