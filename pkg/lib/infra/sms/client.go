package sms

import (
	"context"
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
	TemplateName      string
	LanguageTag       string
	TemplateVariables *TemplateVariables
}

type RawClient interface {
	Send(ctx context.Context, opts SendOptions) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("sms-client")} }

type Client struct {
	Logger                       Logger
	DevMode                      config.DevMode
	MessagingConfig              *config.MessagingConfig
	FeatureTestModeSMSSuppressed config.FeatureTestModeSMSSuppressed
	TestModeSMSConfig            *config.TestModeSMSConfig
	ClientResolver               *ClientResolver
}

func (c *Client) Send(ctx context.Context, opts SendOptions) error {
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
			WithField("template_name", opts.TemplateName).
			WithField("language_tag", opts.LanguageTag).
			WithField("template_variables", opts.TemplateVariables).
			Warn("skip sending SMS in development mode")
		return nil
	}

	client, _, err := c.ClientResolver.ResolveClient()

	if err != nil {
		return err
	}

	return client.Send(ctx, opts)
}

func (c *Client) testModeSend(opts SendOptions) {
	c.Logger.
		WithField("recipient", opts.To).
		WithField("sender", opts.Sender).
		WithField("body", opts.Body).
		Warn("sending SMS is suppressed in test mode")
}
