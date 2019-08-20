package hook

import (
	"net/url"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type mockDeliverer struct {
	WillDeliverFunc       func(eventType event.Type) bool
	DeliveryError         error
	OnDeliverBeforeEvents func(event *event.Event, user *model.User)
	BeforeEvents          []mockDelivererBeforeEvent
	NonBeforeEvents       []mockDelivererNonBeforeEvent
}

type mockDelivererBeforeEvent struct {
	Event *event.Event
	User  *model.User
}

type mockDelivererNonBeforeEvent struct {
	Event   *event.Event
	Timeout time.Duration
}

func newMockDeliverer() *mockDeliverer {
	return &mockDeliverer{}
}

func (deliverer *mockDeliverer) WillDeliver(eventType event.Type) bool {
	if deliverer.WillDeliverFunc == nil {
		return true
	}
	return deliverer.WillDeliverFunc(eventType)
}

func (deliverer *mockDeliverer) DeliverBeforeEvent(baseURL *url.URL, event *event.Event, user *model.User) error {
	_event := *event
	_user := *user
	deliverer.BeforeEvents = append(deliverer.BeforeEvents, mockDelivererBeforeEvent{
		Event: &_event,
		User:  &_user,
	})
	if deliverer.OnDeliverBeforeEvents != nil {
		deliverer.OnDeliverBeforeEvents(event, user)
	}
	return deliverer.DeliveryError
}

func (deliverer *mockDeliverer) DeliverNonBeforeEvent(baseURL *url.URL, event *event.Event, timeout time.Duration) error {
	_event := *event
	deliverer.NonBeforeEvents = append(deliverer.NonBeforeEvents, mockDelivererNonBeforeEvent{
		Event:   &_event,
		Timeout: timeout,
	})
	return deliverer.DeliveryError
}

var _ Deliverer = &mockDeliverer{}
