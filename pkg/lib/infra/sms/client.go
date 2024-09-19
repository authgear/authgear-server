package sms

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrNoAvailableClient = errors.New("no available SMS client")
var ErrAmbiguousClient = errors.New("ambiguous SMS client")

type SendOptions struct {
	Sender            string
	To                string
	Body              string
	AppID             string
	MessageType       string
	TemplateName      string
	LanguageTag       string
	TemplateVariables *TemplateVariables
}

type RawClient interface {
	Send(opts SendOptions) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("sms-client")} }

type Client struct {
	Logger                       Logger
	DevMode                      config.DevMode
	MessagingConfig              *config.MessagingConfig
	FeatureTestModeSMSSuppressed config.FeatureTestModeSMSSuppressed
	TestModeSMSConfig            *config.TestModeSMSConfig
	TwilioClient                 *TwilioClient
	NexmoClient                  *NexmoClient
	CustomClient                 *CustomClient
}

func (c *Client) Send(opts SendOptions) error {
	if c.FeatureTestModeSMSSuppressed {
		c.testModeSend(opts)
		return nil
	}

	if c.TestModeSMSConfig.Enabled {
		if r, ok := c.TestModeSMSConfig.MatchTarget(opts.To); ok && r.Suppressed {
			c.testModeSend(opts)
			return nil
		}
	}

	if c.DevMode {
		c.Logger.
			WithField("recipient", opts.To).
			WithField("sender", opts.Sender).
			WithField("body", opts.Body).
			WithField("app_id", opts.AppID).
			WithField("message_type", opts.MessageType).
			WithField("template_name", opts.TemplateName).
			WithField("language_tag", opts.LanguageTag).
			WithField("template_variables", opts.TemplateVariables).
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
		client = c.CustomClient
	default:
		var availableClients []RawClient = []RawClient{}
		if c.NexmoClient != nil {
			availableClients = append(availableClients, c.NexmoClient)
		}
		if c.TwilioClient != nil {
			availableClients = append(availableClients, c.TwilioClient)
		}
		if c.CustomClient != nil {
			availableClients = append(availableClients, c.CustomClient)
		}
		if len(availableClients) == 0 {
			return ErrNoAvailableClient
		}
		if len(availableClients) > 1 {
			return ErrAmbiguousClient
		}
		client = availableClients[0]
	}

	return client.Send(opts)
}

func (c *Client) testModeSend(opts SendOptions) {
	c.Logger.
		WithField("recipient", opts.To).
		WithField("sender", opts.Sender).
		WithField("body", opts.Body).
		Warn("sending SMS is suppressed in test mode")
}
