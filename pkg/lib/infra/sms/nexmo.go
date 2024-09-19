package sms

import (
	"errors"
	"fmt"

	nexmo "github.com/njern/gonexmo"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrMissingNexmoConfiguration = errors.New("nexmo: configuration is missing")

type NexmoClient struct {
	NexmoClient *nexmo.Client
}

func NewNexmoClient(c *config.NexmoCredentials) *NexmoClient {
	if c == nil {
		return nil
	}

	nexmoClient, _ := nexmo.NewClient(c.APIKey, c.APISecret)
	return &NexmoClient{
		NexmoClient: nexmoClient,
	}
}

func (n *NexmoClient) Send(opts SendOptions) error {
	if n.NexmoClient == nil {
		return ErrMissingNexmoConfiguration
	}

	message := nexmo.SMSMessage{
		From:  opts.Sender,
		To:    opts.To,
		Type:  nexmo.Text,
		Text:  opts.Body,
		Class: nexmo.Standard,
	}

	resp, err := n.NexmoClient.SMS.Send(&message)
	if err != nil {
		return fmt.Errorf("nexmo: %w", err)
	}

	if resp.MessageCount == 0 {
		err = errors.New("nexmo: no sms is sent")
		return err
	}

	report := resp.Messages[0]
	if report.ErrorText != "" {
		err = fmt.Errorf("nexmo: %s", report.ErrorText)
		return err
	}

	return nil
}
