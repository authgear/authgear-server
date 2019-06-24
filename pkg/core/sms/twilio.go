package sms

import (
	"errors"

	"github.com/sfreiberg/gotwilio"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type TwilioClient struct {
	From string
	*gotwilio.Twilio
}

func NewTwilioClient(c config.NewTwilioConfiguration) *TwilioClient {
	if c.AccountSID == "" || c.AuthToken == "" {
		panic(errors.New("Twilio account sid or auth token is empty"))
	}

	return &TwilioClient{
		From:   c.From,
		Twilio: gotwilio.NewTwilioClient(c.AccountSID, c.AuthToken),
	}
}

func (t *TwilioClient) Send(to string, body string) error {
	_, exception, err := t.SendSMS(t.From, to, body, "", "")
	if err != nil {
		return err
	}

	if exception != nil {
		err = errors.New(exception.Message)
		return err
	}

	return nil
}
