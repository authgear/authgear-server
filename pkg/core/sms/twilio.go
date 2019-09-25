package sms

import (
	"errors"

	"github.com/sfreiberg/gotwilio"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

var ErrMissingTwilioConfiguration = errors.New("missing twilio configuration")

type TwilioClient struct {
	From         string
	TwilioClient *gotwilio.Twilio
}

func NewTwilioClient(c config.TwilioConfiguration) *TwilioClient {
	var twilioClient *gotwilio.Twilio
	if c.AccountSID != "" && c.AuthToken != "" {
		twilioClient = gotwilio.NewTwilioClient(c.AccountSID, c.AuthToken)
	}

	return &TwilioClient{
		From:         c.From,
		TwilioClient: twilioClient,
	}
}

func (t *TwilioClient) Send(to string, body string) error {
	if t.TwilioClient == nil {
		return ErrMissingTwilioConfiguration
	}
	_, exception, err := t.TwilioClient.SendSMS(t.From, to, body, "", "")
	if err != nil {
		return err
	}

	if exception != nil {
		err = errors.New(exception.Message)
		return err
	}

	return nil
}
