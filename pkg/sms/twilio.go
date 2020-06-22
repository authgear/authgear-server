package sms

import (
	"github.com/sfreiberg/gotwilio"
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var ErrMissingTwilioConfiguration = errors.New("twilio: configuration is missing")

type TwilioClient struct {
	TwilioClient *gotwilio.Twilio
}

func NewTwilioClient(c *config.TwilioCredentials) *TwilioClient {
	var twilioClient *gotwilio.Twilio
	if c != nil {
		twilioClient = gotwilio.NewTwilioClient(c.AccountSID, c.AuthToken)
	}

	return &TwilioClient{
		TwilioClient: twilioClient,
	}
}

func (t *TwilioClient) Send(from string, to string, body string) error {
	if t.TwilioClient == nil {
		return ErrMissingTwilioConfiguration
	}
	_, exception, err := t.TwilioClient.SendSMS(from, to, body, "", "")
	if err != nil {
		return errors.Newf("twilio: %w", err)
	}

	if exception != nil {
		err = errors.Newf("twilio: %s", exception.Message)
		return err
	}

	return nil
}
