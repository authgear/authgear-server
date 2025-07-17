package sms

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

type Sender struct {
	ClientResolver *ClientResolver
}

func (c *Sender) Send(ctx context.Context, client smsapi.Client, opts smsapi.SendOptions) error {
	return client.Send(ctx, opts)
}

func (c *Sender) ResolveClient() (smsapi.Client, error) {
	client, _, err := c.ClientResolver.ResolveClient()

	if err != nil {
		return nil, err
	}

	return client, err
}
