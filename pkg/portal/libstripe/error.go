package libstripe

import (
	"errors"
)

var ErrUnknownEvent = errors.New("unknown stripe event")
var ErrCustomerAlreadySubscribed = errors.New("custom already subscribed")
var ErrAppAlreadySubscribed = errors.New("app already subscribed")
