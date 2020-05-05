package sms

import (
	"context"
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/intl"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type RawClient interface {
	Send(from string, to string, body string) error
}

type ClientImpl struct {
	LocalizationConfiguration *config.LocalizationConfiguration
	RawClient                 RawClient
	Context                   context.Context
}

func NewClient(ctx context.Context, appConfig *config.AppConfiguration) Client {
	var client RawClient

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

	return &ClientImpl{
		LocalizationConfiguration: appConfig.Localization,
		RawClient:                 client,
		Context:                   ctx,
	}
}

func (c *ClientImpl) Send(opts SendOptions) error {
	if c.RawClient != nil {
		tags := intl.GetPreferredLanguageTags(c.Context)
		from := intl.LocalizeStringMap(tags, intl.Fallback(c.LocalizationConfiguration.FallbackLanguage), opts.MessageConfig, "sender")
		return c.RawClient.Send(from, opts.To, opts.Body)
	}
	return ErrNoAvailableClient
}
