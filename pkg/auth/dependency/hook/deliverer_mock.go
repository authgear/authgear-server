package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type MockDeliverer struct {
	DeliveryError error
	BeforeEvents  []MockDelivererBeforeEvent
}

type MockDelivererBeforeEvent struct {
	Event *event.Event
	User  *model.User
}

func NewMockDeliverer() *MockDeliverer {
	return &MockDeliverer{}
}

func (deliverer *MockDeliverer) Reset() {
	*deliverer = MockDeliverer{}
}

func (deliverer *MockDeliverer) DeliverBeforeEvent(event *event.Event, user *model.User) error {
	deliverer.BeforeEvents = append(deliverer.BeforeEvents, MockDelivererBeforeEvent{
		Event: event,
		User:  user,
	})
	return deliverer.DeliveryError
}

var _ Deliverer = &MockDeliverer{}
