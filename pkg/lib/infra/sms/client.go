package sms

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type SendOptions struct {
	MessageConfig config.SMSMessageConfig
	To            string
	Body          string
}

type RawClient interface {
	Send(from string, to string, body string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("sms-client")} }

type Client struct {
	Context            context.Context
	Logger             Logger
	ServerConfig       *config.ServerConfig
	MessagingConfig    *config.MessagingConfig
	LocalizationConfig *config.LocalizationConfig
	TwilioClient       *TwilioClient
	NexmoClient        *NexmoClient
}

func (c *Client) Send(opts SendOptions) error {
	if c.ServerConfig.DevMode {
		c.Logger.
			WithField("recipient", opts.To).
			WithField("body", opts.Body).
			Warn("skip sending SMS in development mode")
		return nil
	}

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
	default:
		return ErrNoAvailableClient
	}

	tags := intl.GetPreferredLanguageTags(c.Context)
	from := intl.LocalizeStringMap(tags, intl.Fallback(c.LocalizationConfig.FallbackLanguage), opts.MessageConfig, "sender")
	return client.Send(from, opts.To, opts.Body)
}
