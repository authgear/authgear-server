package libstripe

import (
	"github.com/stripe/stripe-go/v72"
)

type EventType string

const (
	EventTypeCheckoutSessionCompleted    EventType = "checkout.session.completed"
	EventTypeCustomerSubscriptionCreated EventType = "customer.subscription.created"
	EventTypeCustomerSubscriptionUpdated EventType = "customer.subscription.updated"
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

func (e *CustomerSubscriptionEvent) IsSubscriptionActive() bool {
	return e.StripeSubscriptionStatus == stripe.SubscriptionStatusActive
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

var _ Event = &CheckoutSessionCompletedEvent{}
var _ Event = &CustomerSubscriptionCreatedEvent{}
var _ Event = &CustomerSubscriptionUpdatedEvent{}
