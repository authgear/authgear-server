package hook

import (
	"fmt"
	"testing"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDispatchEvent(t *testing.T) {
	requestID := "request-id"
	timeProvider := time.MockProvider{TimeNowUTC: gotime.Date(2006, 1, 2, 15, 4, 5, 0, gotime.UTC)}
	store := newMockStore()
	authContext := auth.NewMockContextGetter()
	deliverer := newMockDeliverer()
	authInfoStore := authinfo.NewMockStore()
	userProfileStore := userprofile.NewMockUserProfileStore()

	provider := NewProvider(
		requestID,
		store,
		authContext,
		&timeProvider,
		authInfoStore,
		userProfileStore,
		deliverer,
	).(*providerImpl)

	reset := func() {
		store.Reset()
		deliverer.Reset()
		authContext.Set(nil, nil)
		authInfoStore.AuthInfoMap = map[string]authinfo.AuthInfo{}
		userProfileStore.Data = map[string]map[string]interface{}{}
		provider.PersistentEventPayloads = nil
	}

	Convey("Dispatch operation events", t, func() {
		user := model.User{
			ID: "user-id",
		}
		identity := model.Identity{
			ID: "principal-id",
		}
		payload := event.SessionCreateEvent{
			Reason:   event.SessionCreateReasonLogin,
			User:     user,
			Identity: identity,
		}

		Convey("should be successful", func() {
			reset()

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
				mockDelivererBeforeEvent{
					Event: &event.Event{
						ID:      deliverer.BeforeEvents[0].Event.ID,
						Type:    event.BeforeSessionCreate,
						Version: 2,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:   1136214245,
							RequestID:   &requestID,
							UserID:      nil,
							PrincipalID: nil,
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
			reset()
			deliverer.WillDeliverFunc = func(eventType event.Type) bool {
				return false
			}

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{})
			So(store.nextSequenceNumber, ShouldEqual, 1)
			So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
				payload,
			})
		})

		Convey("should use mutated payload", func() {
			reset()
			deliverer.OnDeliverBeforeEvents = func(ev *event.Event, user *model.User) {
				payload := ev.Payload.(event.SessionCreateEvent)
				payload.Reason = event.SessionCreateReasonSignup
				ev.Payload = payload
			}

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.NonBeforeEvents, ShouldResemble, []mockDelivererNonBeforeEvent{})
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
				mockDelivererBeforeEvent{
					Event: &event.Event{
						ID:      deliverer.BeforeEvents[0].Event.ID,
						Type:    event.BeforeSessionCreate,
						Version: 2,
						Seq:     1,
						Payload: payload,
						Context: event.Context{
							Timestamp:   1136214245,
							RequestID:   &requestID,
							UserID:      nil,
							PrincipalID: nil,
						},
					},
					User: &user,
				},
			})
			So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
				event.SessionCreateEvent{
					Reason:   event.SessionCreateReasonSignup,
					User:     user,
					Identity: identity,
				},
			})
		})

		Convey("should include auth info", func() {
			reset()
			userID := "user-id"
			principalID := "principal-id"
			authContext.Set(
				&authinfo.AuthInfo{ID: userID},
				&authtoken.Token{PrincipalID: principalID},
			)

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents[0].Event.Context, ShouldResemble, event.Context{
				Timestamp:   1136214245,
				RequestID:   &requestID,
				UserID:      &userID,
				PrincipalID: &principalID,
			})
		})

		Convey("should return delivery error", func() {
			store.Reset()
			deliverer.Reset()
			deliverer.DeliveryError = fmt.Errorf("Failed to deliver")
			authContext.Set(nil, nil)
			provider.PersistentEventPayloads = nil

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeError, "Failed to deliver")
		})
	})

	Convey("Dispatch notification events", t, func() {
		user := model.User{
			ID: "user-id",
		}
		payload := event.UserSyncEvent{
			User: user,
		}

		Convey("should be successful", func() {
			reset()

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.NonBeforeEvents, ShouldResemble, []mockDelivererNonBeforeEvent{})
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{})
			So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
				payload,
			})
		})
	})

	Convey("When transaction is about to commit", t, func() {
		Convey("should generate & persist events", func() {
			reset()
			provider.PersistentEventPayloads = []event.Payload{
				event.SessionCreateEvent{
					User: model.User{
						ID: "user-id",
					},
				},
			}
			authInfoStore.AuthInfoMap = map[string]authinfo.AuthInfo{
				"user-id": authinfo.AuthInfo{
					ID:         "user-id",
					VerifyInfo: map[string]bool{"user@example.com": true},
				},
			}
			userProfileStore.Data = map[string]map[string]interface{}{
				"user-id": map[string]interface{}{"user": true},
			}

			err := provider.WillCommitTx()

			So(err, ShouldBeNil)
			So(provider.PersistentEventPayloads, ShouldBeNil)
			So(store.persistedEvents, ShouldResemble, []*event.Event{
				&event.Event{
					ID:      store.persistedEvents[0].ID,
					Type:    event.AfterSessionCreate,
					Version: 2,
					Seq:     1,
					Payload: event.SessionCreateEvent{
						User: model.User{
							ID: "user-id",
						},
					},
					Context: event.Context{
						Timestamp:   1136214245,
						RequestID:   &requestID,
						UserID:      nil,
						PrincipalID: nil,
					},
				},
				&event.Event{
					ID:      store.persistedEvents[1].ID,
					Type:    event.UserSync,
					Version: 2,
					Seq:     2,
					Payload: event.UserSyncEvent{
						User: model.User{
							ID:         "user-id",
							VerifyInfo: map[string]bool{"user@example.com": true},
							Metadata:   map[string]interface{}{"user": true},
						},
					},
					Context: event.Context{
						Timestamp:   1136214245,
						RequestID:   &requestID,
						UserID:      nil,
						PrincipalID: nil,
					},
				},
			})
		})

		Convey("should not generate events that would not be delivered", func() {
			reset()
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
			authInfoStore.AuthInfoMap = map[string]authinfo.AuthInfo{
				"user-id": authinfo.AuthInfo{
					ID:         "user-id",
					VerifyInfo: map[string]bool{"user@example.com": true},
				},
			}
			userProfileStore.Data = map[string]map[string]interface{}{
				"user-id": map[string]interface{}{"user": true},
			}

			err := provider.WillCommitTx()

			So(err, ShouldBeNil)
			So(provider.PersistentEventPayloads, ShouldBeNil)
			So(store.nextSequenceNumber, ShouldEqual, 2)
			So(store.persistedEvents, ShouldResemble, []*event.Event{
				&event.Event{
					ID:      store.persistedEvents[0].ID,
					Type:    event.UserSync,
					Version: 2,
					Seq:     1,
					Payload: event.UserSyncEvent{
						User: model.User{
							ID:         "user-id",
							VerifyInfo: map[string]bool{"user@example.com": true},
							Metadata:   map[string]interface{}{"user": true},
						},
					},
					Context: event.Context{
						Timestamp:   1136214245,
						RequestID:   &requestID,
						UserID:      nil,
						PrincipalID: nil,
					},
				},
			})
		})
	})
}
