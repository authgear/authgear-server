package sms

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrMissingCustomSMSProviderConfiguration = errors.New("sms: custom provider configuration is missing")

type CustomClient struct {
	Config *config.CustomSMSProviderConfigs
}

func NewCustomClient(c *config.CustomSMSProviderConfigs) *CustomClient {
	if c == nil {
		return nil
	}

	return &CustomClient{}
}

func (t *CustomClient) Send(from string, to string, body string) error {
	if t.Config == nil {
		return ErrMissingCustomSMSProviderConfiguration
	}

	return nil
}
