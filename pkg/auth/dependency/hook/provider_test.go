package hook

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/log"

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
		ctx := context.Background()

		provider := &Provider{
			Context:   ctx,
			Logger:    Logger{log.Null},
			DBContext: db.NewMockTxContext(),
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

		Convey("dispatching operation events", func() {
			user := model.User{
				ID: "user-id",
			}
			identity := model.Identity{
				Type: "login_id",
			}
			payload := event.SessionCreateEvent{
				Reason:   "login",
				User:     user,
				Identity: identity,
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
					&user,
				).Return(nil)

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})

			Convey("should not generate before events that would not be delivered", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(false)

				err := provider.DispatchEvent(
					payload,
					&user,
				)

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
					&user,
				).DoAndReturn(func(ev *event.Event, user *model.User) error {
					payload := ev.Payload.(event.SessionCreateEvent)
					payload.Reason = "signup"
					ev.Payload = payload
					return nil
				})

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					event.SessionCreateEvent{
						Reason:   "signup",
						User:     user,
						Identity: identity,
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
					&authn.UserInfo{
						ID: userID,
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
					&user,
				).Return(nil)

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
			})

			Convey("should return delivery error", func() {
				deliverer.EXPECT().WillDeliver(event.BeforeSessionCreate).Return(true)
				deliverer.EXPECT().DeliverBeforeEvent(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("failed to deliver"))

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeError, "failed to dispatch event")
			})
		})

		Convey("dispatching notification events", func() {
			user := model.User{
				ID: "user-id",
			}
			payload := event.UserSyncEvent{
				User: user,
			}

			Convey("should be successful", func() {
				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(provider.persistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})
		})

		Convey("when transaction is about to commit", func() {
			Convey("should generate & persist events", func() {
				provider.persistentEventPayloads = []event.Payload{
					event.SessionCreateEvent{
						User: model.User{
							ID: "user-id",
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					ID: "user-id",
					Metadata: map[string]interface{}{
						"user": true,
					},
				}, nil)
				deliverer.EXPECT().WillDeliver(event.UserSync).Return(true)
				deliverer.EXPECT().WillDeliver(event.AfterSessionCreate).Return(true)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:   "0000000000000001",
						Type: event.AfterSessionCreate,
						Seq:  1,
						Payload: event.SessionCreateEvent{
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
						Payload: event.UserSyncEvent{
							User: model.User{
								ID:       "user-id",
								Metadata: map[string]interface{}{"user": true},
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
					event.SessionCreateEvent{
						User: model.User{
							ID: "user-id",
						},
					},
				}
				users.EXPECT().Get("user-id").Return(&model.User{
					ID: "user-id",
					Metadata: map[string]interface{}{
						"user": true,
					},
				}, nil)
				deliverer.EXPECT().WillDeliver(event.UserSync).Return(true)
				deliverer.EXPECT().WillDeliver(event.AfterSessionCreate).Return(false)
				store.EXPECT().AddEvents([]*event.Event{
					{
						ID:   "0000000000000001",
						Type: event.UserSync,
						Seq:  1,
						Payload: event.UserSyncEvent{
							User: model.User{
								ID:       "user-id",
								Metadata: map[string]interface{}{"user": true},
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
