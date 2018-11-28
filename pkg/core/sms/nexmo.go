package sms

import (
	"errors"

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

func (n *NexmoClient) Send(to string, body string) error {
	message := nexmo.SMSMessage{
		From:  n.From,
		To:    to,
		Type:  nexmo.Text,
		Text:  body,
		Class: nexmo.Standard,
	}

	resp, err := n.SMS.Send(&message)
	if err != nil {
		return err
	}

	if resp.MessageCount == 0 {
		err = errors.New("Unable to send sms")
		return err
	}

	return nil
}
