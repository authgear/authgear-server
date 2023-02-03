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
	Config     *config.CustomSMSProviderConfig
	DenoHook   hook.DenoHook
	SMSWebHook SMSWebHook
}

func NewCustomClient(c *config.CustomSMSProviderConfig, d hook.DenoHook, w SMSWebHook) *CustomClient {
	if c == nil {
		return nil
	}

	return &CustomClient{
		Config:     c,
		DenoHook:   d,
		SMSWebHook: w,
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
	payload := SendSMSPayload{To: to, Body: body}
	switch {
	case c.DenoHook.SupportURL(u):
		_, err := c.DenoHook.RunSync(u, &SendSMSPayload{To: to, Body: body}, timeout)
		return err
	case c.SMSWebHook.SupportURL(u):
		return c.SMSWebHook.Call(u, payload)
	default:
		return fmt.Errorf("unsupported hook URL: %v", u)
	}
}

type SMSWebHook struct {
	hook.WebHook
	SyncHTTP HookHTTPClient
}

func (w *SMSWebHook) Call(u *url.URL, payload SendSMSPayload) error {
	req, err := w.PrepareRequest(u, payload)
	if err != nil {
		return err
	}
	return w.PerformNoResponse(w.SyncHTTP.Client, req)
}
