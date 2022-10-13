package libstripe

import (
	"github.com/stripe/stripe-go/v72"
)

type EventType string

const (
	EventTypeCheckoutSessionCompleted    EventType = "checkout.session.completed"
	EventTypeCustomerSubscriptionCreated EventType = "customer.subscription.created"
	EventTypeCustomerSubscriptionUpdated EventType = "customer.subscription.updated"
	EventTypeCustomerSubscriptionDeleted EventType = "customer.subscription.deleted"
)

// type StripeSubscription

type Event interface {
	EventType() EventType
}

type CheckoutSessionCompletedEvent struct {
	AppID                   string
	PlanName                string
	StripeCustomerID        string
	StripeCheckoutSessionID string
}

func (e *CheckoutSessionCompletedEvent) EventType() EventType {
	return EventTypeCheckoutSessionCompleted
}

type CustomerSubscriptionEvent struct {
	StripeSubscriptionID     string
	StripeCustomerID         string
	AppID                    string
	PlanName                 string
	StripeSubscriptionStatus stripe.SubscriptionStatus
}

func (e *CustomerSubscriptionEvent) IsSubscriptionIncompleteExpired() bool {
	return e.StripeSubscriptionStatus == stripe.SubscriptionStatusIncompleteExpired
}

func (e *CustomerSubscriptionEvent) IsSubscriptionActive() bool {
	return e.StripeSubscriptionStatus == stripe.SubscriptionStatusActive
}

func (e *CustomerSubscriptionEvent) IsSubscriptionCanceled() bool {
	return e.StripeSubscriptionStatus == stripe.SubscriptionStatusCanceled
}

type CustomerSubscriptionCreatedEvent struct {
	*CustomerSubscriptionEvent
}

func (e *CustomerSubscriptionCreatedEvent) EventType() EventType {
	return EventTypeCustomerSubscriptionCreated
}

type CustomerSubscriptionUpdatedEvent struct {
	*CustomerSubscriptionEvent
}

func (e *CustomerSubscriptionUpdatedEvent) EventType() EventType {
	return EventTypeCustomerSubscriptionUpdated
}

type CustomerSubscriptionDeletedEvent struct {
	*CustomerSubscriptionEvent
}

func (e *CustomerSubscriptionDeletedEvent) EventType() EventType {
	return EventTypeCustomerSubscriptionDeleted
}

var _ Event = &CheckoutSessionCompletedEvent{}
var _ Event = &CustomerSubscriptionCreatedEvent{}
var _ Event = &CustomerSubscriptionUpdatedEvent{}
var _ Event = &CustomerSubscriptionDeletedEvent{}
