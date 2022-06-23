package libstripe

type EventType string

const (
	EventTypeCheckoutSessionCompleted EventType = "checkout.session.completed"
)

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

var _ Event = &CheckoutSessionCompletedEvent{}
