package sms

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("sms-client")} }

type Client struct {
	Logger         Logger
	ClientResolver *ClientResolver
}

var _ smsapi.Client = (*Client)(nil)

func (c *Client) Send(ctx context.Context, opts smsapi.SendOptions) error {
	client, _, err := c.ClientResolver.ResolveClient()

	if err != nil {
		return err
	}

	return client.Send(ctx, opts)
}
