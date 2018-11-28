package sms

import (
	nexmo "github.com/njern/gonexmo"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type NexmoClient struct {
	From string
	*nexmo.Client
}

func NewNexmoClient(c config.NexmoConfiguration) *NexmoClient {
	client, err := nexmo.NewClient(c.APIKey, c.AuthToken)
	if err != nil {
		return nil
	}

	return &NexmoClient{
		From:   c.From,
		Client: client,
	}
}
