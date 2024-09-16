package sms

import (
	"errors"
	"fmt"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrMissingTwilioConfiguration = errors.New("twilio: configuration is missing")

type TwilioClient struct {
	TwilioClient        *twilio.RestClient
	MessagingServiceSID string
}

func NewTwilioClient(c *config.TwilioCredentials) *TwilioClient {
	if c == nil {
		return nil
	}

	return &TwilioClient{
		TwilioClient: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: c.AccountSID,
			Password: c.AuthToken,
		}),
		MessagingServiceSID: c.MessagingServiceSID,
	}
}

func (t *TwilioClient) Send(opts SendOptions) error {
	if t.TwilioClient == nil {
		return ErrMissingTwilioConfiguration
	}

	params := &api.CreateMessageParams{}
	params.SetBody(opts.Body)
	params.SetTo(opts.To)
	if t.MessagingServiceSID != "" {
		params.SetMessagingServiceSid(t.MessagingServiceSID)
	} else {
		params.SetFrom(opts.Sender)
	}

	_, err := t.TwilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio: %w", err)
	}

	return nil
}
