package hook

import (
	"fmt"
	"testing"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
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

	provider := NewProvider(requestID, store, authContext, &timeProvider, deliverer).(*providerImpl)

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
			store.Reset()
			deliverer.Reset()
			deliverer.DeliveryError = nil
			authContext.Set(nil, nil)
			provider.PersistentEventPayloads = nil

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
				mockDelivererBeforeEvent{
					Event: &event.Event{
						ID:         deliverer.BeforeEvents[0].Event.ID,
						Type:       event.BeforeSessionCreate,
						Version:    2,
						SequenceNo: 1,
						Payload:    payload,
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

		Convey("should use mutated payload", func() {
			store.Reset()
			deliverer.Reset()
			deliverer.DeliveryError = nil
			deliverer.OnDeliverBeforeEvents = func(ev *event.Event, user *model.User) {
				payload := ev.Payload.(event.SessionCreateEvent)
				payload.Reason = event.SessionCreateReasonSignup
				ev.Payload = payload
			}
			authContext.Set(nil, nil)
			provider.PersistentEventPayloads = nil

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{
				mockDelivererBeforeEvent{
					Event: &event.Event{
						ID:         deliverer.BeforeEvents[0].Event.ID,
						Type:       event.BeforeSessionCreate,
						Version:    2,
						SequenceNo: 1,
						Payload:    payload,
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
			store.Reset()
			deliverer.Reset()
			deliverer.DeliveryError = nil
			userID := "user-id"
			principalID := "principal-id"
			authContext.Set(
				&authinfo.AuthInfo{ID: userID},
				&authtoken.Token{PrincipalID: principalID},
			)
			provider.PersistentEventPayloads = nil

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
			store.Reset()
			deliverer.Reset()
			deliverer.DeliveryError = nil
			authContext.Set(nil, nil)
			provider.PersistentEventPayloads = nil

			err := provider.DispatchEvent(
				payload,
				&user,
			)

			So(err, ShouldBeNil)
			So(deliverer.BeforeEvents, ShouldResemble, []mockDelivererBeforeEvent{})
			So(provider.PersistentEventPayloads, ShouldResemble, []event.Payload{
				payload,
			})
		})
	})
}
