package sms

import (
	"errors"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

var ErrNoAvailableClient = errors.New("no available client")

type clientImpl struct {
	appConfig    config.AppConfiguration
	nexmoClient  *NexmoClient
	twilioClient *TwilioClient
}

func NewClient(appConfig config.AppConfiguration) Client {
	nexmoConfig := appConfig.Nexmo
	twilioConfig := appConfig.Twilio

	var nexmoClient *NexmoClient
	if nexmoConfig.APIKey != "" && nexmoConfig.APISecret != "" {
		nexmoClient = NewNexmoClient(nexmoConfig)
	}

	var twilioClient *TwilioClient
	if twilioConfig.AccountSID != "" && twilioConfig.AuthToken != "" {
		twilioClient = NewTwilioClient(twilioConfig)
	}

	return &clientImpl{
		appConfig:    appConfig,
		nexmoClient:  nexmoClient,
		twilioClient: twilioClient,
	}
}

func (c *clientImpl) Send(to string, body string) error {
	if c.nexmoClient != nil {
		return c.nexmoClient.Send(to, body)
	}
	if c.twilioClient != nil {
		return c.twilioClient.Send(to, body)
	}
	return ErrNoAvailableClient
}
