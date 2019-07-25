package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type mockDeliverer struct {
	DeliveryError         error
	OnDeliverBeforeEvents func(event *event.Event, user *model.User)
	BeforeEvents          []mockDelivererBeforeEvent
}

type mockDelivererBeforeEvent struct {
	Event *event.Event
	User  *model.User
}

func newMockDeliverer() *mockDeliverer {
	return &mockDeliverer{
		BeforeEvents: []mockDelivererBeforeEvent{},
	}
}

func (deliverer *mockDeliverer) Reset() {
	*deliverer = *newMockDeliverer()
}

func (deliverer *mockDeliverer) DeliverBeforeEvent(event *event.Event, user *model.User) error {
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

var _ Deliverer = &mockDeliverer{}
