package model

type Subscription struct {
	ID                      string
	AppID                   string
	StripeCheckoutSessionID string
	StripeCustomerID        string
	StripeSubscriptionID    string
}
