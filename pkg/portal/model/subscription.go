package model

import (
	"time"
)

type SubscriptionCheckoutStatus string

const (
	SubscriptionCheckoutStatusOpen       SubscriptionCheckoutStatus = "open"
	SubscriptionCheckoutStatusCompleted  SubscriptionCheckoutStatus = "completed"
	SubscriptionCheckoutStatusSubscribed SubscriptionCheckoutStatus = "subscribed"
)

type Subscription struct {
	ID                   string
	AppID                string
	StripeSubscriptionID string
	StripeCustomerID     string
}

type SubscriptionCheckout struct {
	ID                      string
	AppID                   string
	StripeCheckoutSessionID string
	StripeCustomerID        *string
	Status                  SubscriptionCheckoutStatus
	ExpireAt                time.Time
}
