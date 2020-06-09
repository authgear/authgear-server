package hook

import (
	"context"
	"fmt"
	"testing"
	gotime "time"

	gomock "github.com/golang/mock/gomock"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDispatchEvent(t *testing.T) {
	Convey("Hook Provider", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		timeProvider := time.MockProvider{TimeNowUTC: gotime.Date(2006, 1, 2, 15, 4, 5, 0, gotime.UTC)}
		store := newMockStore()
		deliverer := newMockDeliverer()
		users := NewMockUserProvider(ctrl)
		ctx := context.Background()

		provider := NewProvider(
			ctx,
			store,
			db.NewMockTxContext(),
			&timeProvider,
			users,
			deliverer,
			logging.NewNullFactory(),
		).(*providerImpl)

		provider.PersistentEventPayloads = nil

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
				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
					mockDelivererBeforeEvent{
						Event: &event.Event{
							ID:      "0000000000000001",
							Type:    event.BeforeSessionCreate,
							Seq:     1,
							Payload: payload,
							Context: event.Context{
								Timestamp: 1136214245,
								UserID:    nil,
							},
						},
						User: &user,
					},
				})
				So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})

			Convey("should not generate before events that would not be delivered", func() {
				deliverer.WillDeliverFunc = func(eventType event.Type) bool {
					return false
				}

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(deliverer.BeforeEvents, ShouldBeEmpty)
				So(store.nextSequenceNumber, ShouldEqual, 1)
				So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})

			Convey("should use mutated payload", func() {
				deliverer.OnDeliverBeforeEvents = func(ev *event.Event, user *model.User) {
					payload := ev.Payload.(event.SessionCreateEvent)
					payload.Reason = "signup"
					ev.Payload = payload
				}

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(deliverer.NonBeforeEvents, ShouldBeEmpty)
				So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
					mockDelivererBeforeEvent{
						Event: &event.Event{
							ID:      "0000000000000001",
							Type:    event.BeforeSessionCreate,
							Seq:     1,
							Payload: payload,
							Context: event.Context{
								Timestamp: 1136214245,
								UserID:    nil,
							},
						},
						User: &user,
					},
				})
				So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
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

				err := provider.DispatchEvent(
					payload,
					&user,
				)

				So(err, ShouldBeNil)
				So(deliverer.BeforeEvents[0].Event.Context, ShouldResemble, event.Context{
					Timestamp: 1136214245,
					UserID:    &userID,
					Session: &model.Session{
						ID: "user-id-principal-id",
					},
				})
			})

			Convey("should return delivery error", func() {
				deliverer.DeliveryError = fmt.Errorf("failed to deliver")

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
				So(deliverer.NonBeforeEvents, ShouldBeEmpty)
				So(deliverer.BeforeEvents, ShouldBeEmpty)
				So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
					payload,
				})
			})
		})

		Convey("when transaction is about to commit", func() {
			Convey("should generate & persist events", func() {
				provider.PersistentEventPayloads = []event.Payload{
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

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.PersistentEventPayloads, ShouldBeNil)
				So(store.persistedEvents, ShouldResemble, []*event.Event{
					&event.Event{
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
					&event.Event{
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
			})

			Convey("should not generate events that would not be delivered", func() {
				deliverer.WillDeliverFunc = func(eventType event.Type) bool {
					return eventType == event.UserSync
				}
				provider.PersistentEventPayloads = []event.Payload{
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

				err := provider.WillCommitTx()

				So(err, ShouldBeNil)
				So(provider.PersistentEventPayloads, ShouldBeNil)
				So(store.nextSequenceNumber, ShouldEqual, 2)
				So(store.persistedEvents, ShouldResemble, []*event.Event{
					&event.Event{
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
			})
		})
	})
}
