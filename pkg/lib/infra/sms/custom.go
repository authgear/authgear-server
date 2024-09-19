package sms

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
)

var ErrMissingCustomSMSProviderConfiguration = errors.New("sms: custom provider configuration is missing")

type CustomClient struct {
	Config      *config.CustomSMSProviderConfig
	SMSDenoHook SMSDenoHook
	SMSWebHook  SMSWebHook
}

func NewCustomClient(c *config.CustomSMSProviderConfig, d SMSDenoHook, w SMSWebHook) *CustomClient {
	if c == nil {
		return nil
	}

	return &CustomClient{
		Config:      c,
		SMSDenoHook: d,
		SMSWebHook:  w,
	}
}

type SendSMSPayload struct {
	To                string             `json:"to"`
	Body              string             `json:"body"`
	AppID             string             `json:"app_id"`
	MessageType       string             `json:"message_type"`
	TemplateName      string             `json:"template_name"`
	LanguageTag       string             `json:"language_tag"`
	TemplateVariables *TemplateVariables `json:"template_variables"`
}

func (c *CustomClient) Send(opts SendOptions) error {
	if c.Config == nil {
		return ErrMissingCustomSMSProviderConfiguration
	}
	u, err := url.Parse(c.Config.URL)
	if err != nil {
		return err
	}
	payload := SendSMSPayload{
		To:                opts.To,
		Body:              opts.Body,
		AppID:             opts.AppID,
		MessageType:       opts.MessageType,
		TemplateName:      opts.TemplateName,
		LanguageTag:       opts.LanguageTag,
		TemplateVariables: opts.TemplateVariables,
	}
	switch {
	case c.SMSDenoHook.SupportURL(u):
		return c.SMSDenoHook.Call(u, payload)
	case c.SMSWebHook.SupportURL(u):
		return c.SMSWebHook.Call(u, payload)
	default:
		return fmt.Errorf("unsupported hook URL: %v", u)
	}
}

type SMSWebHook struct {
	hook.WebHook
	Client HookHTTPClient
}

func (w *SMSWebHook) Call(u *url.URL, payload SendSMSPayload) error {
	req, err := w.PrepareRequest(u, payload)
	if err != nil {
		return err
	}
	return w.PerformNoResponse(w.Client.Client, req)
}

type SMSDenoHook struct {
	hook.DenoHook
	Client HookDenoClient
}

func (d *SMSDenoHook) Call(u *url.URL, payload SendSMSPayload) error {
	return d.RunAsync(d.Client, u, payload)
}
