package sms

import (
	"context"
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/intl"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type RawClient interface {
	Send(from string, to string, body string) error
}

type ClientImpl struct {
	Context            context.Context
	MessagingConfig    *config.MessagingConfig
	LocalizationConfig *config.LocalizationConfig
	TwilioClient       *TwilioClient
	NexmoClient        *NexmoClient
}

func (c *ClientImpl) Send(opts SendOptions) error {
	var client RawClient
	switch c.MessagingConfig.SMSProvider {
	case config.SMSProviderNexmo:
		if c.NexmoClient == nil {
			return ErrNoAvailableClient
		}
		client = c.NexmoClient
	case config.SMSProviderTwilio:
		if c.TwilioClient == nil {
			return ErrNoAvailableClient
		}
		client = c.TwilioClient
	}

	tags := intl.GetPreferredLanguageTags(c.Context)
	from := intl.LocalizeStringMap(tags, intl.Fallback(c.LocalizationConfig.FallbackLanguage), opts.MessageConfig, "sender")
	return client.Send(from, opts.To, opts.Body)
}
