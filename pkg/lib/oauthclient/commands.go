package oauthclient

import (
	"context"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateClient(ctx context.Context, options *NewClientOptions) (*Client, error) {
	client := c.Store.NewClient(options)
	if err := c.Store.CreateClient(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Commands) DeleteClient(ctx context.Context, clientID string) error {
	return c.Store.DeleteClientByClientID(ctx, clientID)
}
