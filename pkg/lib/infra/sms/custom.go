package sms

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

var ErrMissingCustomSMSProviderConfiguration = errors.New("sms: custom provider configuration is missing")

type CustomClientCredentials struct {
	URL     string
	Timeout *config.DurationSeconds
}

func (CustomClientCredentials) smsClientCredentials() {}

type CustomClient struct {
	Config      *config.CustomSMSProviderConfig
	SMSDenoHook SMSDenoHook
	SMSWebHook  SMSWebHook
}

var _ smsapi.Client = (*CustomClient)(nil)

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
	To                string                    `json:"to"`
	Body              string                    `json:"body"`
	AppID             string                    `json:"app_id"`
	TemplateName      string                    `json:"template_name"`
	LanguageTag       string                    `json:"language_tag"`
	TemplateVariables *smsapi.TemplateVariables `json:"template_variables"`
}

func (c *CustomClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
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
		TemplateName:      opts.TemplateName,
		LanguageTag:       opts.LanguageTag,
		TemplateVariables: opts.TemplateVariables,
	}
	switch {
	case c.SMSDenoHook.SupportURL(u):
		return c.SMSDenoHook.Call(ctx, u, payload)
	case c.SMSWebHook.SupportURL(u):
		return c.SMSWebHook.Call(ctx, u, payload)
	default:
		return fmt.Errorf("unsupported hook URL: %v", u)
	}
}

type SMSWebHook struct {
	hook.WebHook
	Client HookHTTPClient
}

func (w *SMSWebHook) Call(ctx context.Context, u *url.URL, payload SendSMSPayload) error {
	// Detach the deadline so that the context is not canceled along with the request.
	ctx = context.WithoutCancel(ctx)
	req, err := w.PrepareRequest(ctx, u, payload)
	if err != nil {
		return err
	}
	return w.PerformNoResponse(w.Client.Client, req)
}

type SMSDenoHook struct {
	hook.DenoHook
	Client HookDenoClient
}

func (d *SMSDenoHook) Call(ctx context.Context, u *url.URL, payload SendSMSPayload) error {
	return d.RunAsync(ctx, d.Client, u, payload)
}
