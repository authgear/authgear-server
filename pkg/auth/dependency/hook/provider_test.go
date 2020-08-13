package hook

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
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
				ID: "user-id",
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
				provider.Context = authn.WithAuthn(
					context.Background(),
					&session.IDPSession{
						ID: "user-id-principal-id",
						Attrs: authn.Attrs{
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
				ID: "user-id",
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
							ID: "user-id",
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					ID: "user-id",
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
								ID: "user-id",
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
								ID: "user-id",
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
							ID: "user-id",
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					ID: "user-id",
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
								ID: "user-id",
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
	})
}
