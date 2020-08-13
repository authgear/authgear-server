package sms

import (
	"errors"
	"fmt"

	"github.com/sfreiberg/gotwilio"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrMissingTwilioConfiguration = errors.New("twilio: configuration is missing")

type TwilioClient struct {
	TwilioClient *gotwilio.Twilio
}

func NewTwilioClient(c *config.TwilioCredentials) *TwilioClient {
	if c == nil {
		return nil
	}

	return &TwilioClient{
		TwilioClient: gotwilio.NewTwilioClient(c.AccountSID, c.AuthToken),
	}
}

func (t *TwilioClient) Send(from string, to string, body string) error {
	if t.TwilioClient == nil {
		return ErrMissingTwilioConfiguration
	}
	_, exception, err := t.TwilioClient.SendSMS(from, to, body, "", "")
	if err != nil {
		return fmt.Errorf("twilio: %w", err)
	}

	if exception != nil {
		err = fmt.Errorf("twilio: %s", exception.Message)
		return err
	}

	return nil
}
