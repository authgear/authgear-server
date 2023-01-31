package sms

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
)

var ErrMissingCustomSMSProviderConfiguration = errors.New("sms: custom provider configuration is missing")

type CustomClient struct {
	Config   *config.CustomSMSProviderConfigs
	DenoHook hook.DenoHook
	WebHook  hook.WebHook
}

func NewCustomClient(c *config.CustomSMSProviderConfigs, d hook.DenoHook, w hook.WebHook) *CustomClient {
	if c == nil {
		return nil
	}

	return &CustomClient{
		Config:   c,
		DenoHook: d,
		WebHook:  w,
	}
}

type SendSMSPayload struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

func (c *CustomClient) Send(from string, to string, body string) error {
	if c.Config == nil {
		return ErrMissingCustomSMSProviderConfiguration
	}
	u, err := url.Parse(c.Config.URL)
	if err != nil {
		return err
	}
	var timeout *time.Duration = nil
	if c.Config.Timeout != nil {
		d := c.Config.Timeout.Duration()
		timeout = &d
	}
	switch {
	case c.DenoHook.SupportURL(u):
		_, err := c.DenoHook.RunSync(u, &SendSMSPayload{To: to, Body: body}, timeout)
		return err
	case c.WebHook.SupportURL(u):
		_, err := c.WebHook.CallSync(u, &SendSMSPayload{To: to, Body: body}, timeout)
		return err
	default:
		return fmt.Errorf("unsupported hook URL: %v", u)
	}
}
