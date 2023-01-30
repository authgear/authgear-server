package sms

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrNoAvailableClient = errors.New("no available SMS client")
var ErrAmbiguousClient = errors.New("ambiguous SMS client")

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
	CustomClient    *CustomClient
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
	case config.SMSProviderCustom:
		if c.CustomClient == nil {
			return ErrNoAvailableClient
		}
	default:
		var availableClients []RawClient = []RawClient{}
		for _, c := range []RawClient{c.NexmoClient, c.TwilioClient, c.CustomClient} {
			if c != nil {
				availableClients = append(availableClients, c)
			}
		}
		if len(availableClients) == 0 {
			return ErrNoAvailableClient
		}
		if len(availableClients) > 1 {
			return ErrAmbiguousClient
		}
		client = availableClients[0]
	}

	return client.Send(opts.Sender, opts.To, opts.Body)
}
