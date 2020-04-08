package sms

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type clientWrapper struct {
	client Client
}

func NewClient(appConfig *config.AppConfiguration) Client {
	var client Client

	switch appConfig.Messages.SMSProvider {
	case config.SMSProviderNexmo:
		nexmoConfig := appConfig.Nexmo
		if nexmoConfig.IsValid() {
			client = NewNexmoClient(nexmoConfig)
		}

	case config.SMSProviderTwilio:
		twilioConfig := appConfig.Twilio
		if twilioConfig.IsValid() {
			client = NewTwilioClient(twilioConfig)
		}
	}

	return &clientWrapper{client}
}

func (c *clientWrapper) Send(from string, to string, body string) error {
	if c.client != nil {
		return c.client.Send(from, to, body)
	}
	return ErrNoAvailableClient
}
