package sms

import (
	"github.com/sfreiberg/gotwilio"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var ErrMissingTwilioConfiguration = errors.New("twilio: configuration is missing")

type TwilioClient struct {
	From         string
	TwilioClient *gotwilio.Twilio
}

func NewTwilioClient(c *config.TwilioConfiguration) *TwilioClient {
	var twilioClient *gotwilio.Twilio
	if c != nil && c.IsValid() {
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
		return errors.Newf("twilio: %w", err)
	}

	if exception != nil {
		err = errors.Newf("twilio: %s", exception.Message)
		return err
	}

	return nil
}
