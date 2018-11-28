package sms

import (
	"github.com/sfreiberg/gotwilio"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type TwilioClient struct {
	From string
	*gotwilio.Twilio
}

func NewTwilioClient(c config.TwilioConfiguration) *TwilioClient {
	return &TwilioClient{
		From:   c.From,
		Twilio: gotwilio.NewTwilioClient(c.AccountSID, c.AuthToken),
	}
}
