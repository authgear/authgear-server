package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type mockDeliverer struct {
	DeliveryError error
	BeforeEvents  []mockDelivererBeforeEvent
}

type mockDelivererBeforeEvent struct {
	Event *event.Event
	User  *model.User
}

func newMockDeliverer() *mockDeliverer {
	return &mockDeliverer{}
}

func (deliverer *mockDeliverer) Reset() {
	*deliverer = mockDeliverer{}
}

func (deliverer *mockDeliverer) DeliverBeforeEvent(event *event.Event, user *model.User) error {
	deliverer.BeforeEvents = append(deliverer.BeforeEvents, mockDelivererBeforeEvent{
		Event: event,
		User:  user,
	})
	return deliverer.DeliveryError
}

var _ Deliverer = &mockDeliverer{}
