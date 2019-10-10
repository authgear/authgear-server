package sms

import (
	"errors"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

var ErrNoAvailableClient = errors.New("no available client")

type clientImpl struct {
	userConfig   config.UserConfiguration
	nexmoClient  *NexmoClient
	twilioClient *TwilioClient
}

func NewClient(userConfig config.UserConfiguration) Client {
	nexmoConfig := userConfig.Nexmo
	twilioConfig := userConfig.Twilio

	var nexmoClient *NexmoClient
	if nexmoConfig.IsValid() {
		nexmoClient = NewNexmoClient(nexmoConfig)
	}

	var twilioClient *TwilioClient
	if twilioConfig.IsValid() {
		twilioClient = NewTwilioClient(twilioConfig)
	}

	return &clientImpl{
		userConfig:   userConfig,
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
