package sms

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type SendOptions struct {
	Sender string
	To     string
	Body   string
}

type RawClient interface {
	Send(from string, to string, body string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("sms-client")} }

type Client struct {
	Logger          Logger
	DevMode         config.DevMode
	MessagingConfig *config.MessagingConfig
	TwilioClient    *TwilioClient
	NexmoClient     *NexmoClient
}

func (c *Client) Send(opts SendOptions) error {
	if c.DevMode {
		c.Logger.
			WithField("recipient", opts.To).
			WithField("sender", opts.Sender).
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

	return client.Send(opts.Sender, opts.To, opts.Body)
}
