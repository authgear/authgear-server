package libstripe

import (
	"errors"
)

var ErrUnknownEvent = errors.New("unknown stripe event")
var ErrCustomerAlreadySubscribed = errors.New("custom already subscribed")
var ErrAppAlreadySubscribed = errors.New("app already subscribed")
var ErrNoSubscription = errors.New("customer has no subscription")
var ErrNoInvoice = errors.New("subscription has no invoice")
var ErrNoPaymentIntent = errors.New("invoice has no payment intent")
