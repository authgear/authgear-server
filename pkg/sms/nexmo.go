package sms

import (
	nexmo "github.com/njern/gonexmo"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/errors"
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

func (n *NexmoClient) Send(from string, to string, body string) error {
	if n.NexmoClient == nil {
		return ErrMissingNexmoConfiguration
	}

	message := nexmo.SMSMessage{
		From:  from,
		To:    to,
		Type:  nexmo.Text,
		Text:  body,
		Class: nexmo.Standard,
	}

	resp, err := n.NexmoClient.SMS.Send(&message)
	if err != nil {
		return errors.Newf("nexmo: %w", err)
	}

	if resp.MessageCount == 0 {
		err = errors.New("nexmo: no sms is sent")
		return err
	}

	report := resp.Messages[0]
	if report.ErrorText != "" {
		err = errors.Newf("nexmo: %s", report.ErrorText)
		return err
	}

	return nil
}
