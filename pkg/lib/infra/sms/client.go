package sms

import (
	"context"
	"errors"

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
	Logger         Logger
	ClientResolver *ClientResolver
}

func (c *Client) Send(ctx context.Context, opts SendOptions) error {
	client, _, err := c.ClientResolver.ResolveClient()

	if err != nil {
		return err
	}

	return client.Send(ctx, opts)
}
