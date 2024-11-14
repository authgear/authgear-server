package custom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

type SMSHookTimeout struct {
	Timeout time.Duration
}

func NewSMSHookTimeout(smsCfg *config.CustomSMSProviderConfig) SMSHookTimeout {
	if smsCfg != nil && smsCfg.Timeout != nil {
		return SMSHookTimeout{Timeout: smsCfg.Timeout.Duration()}
	} else {
		return SMSHookTimeout{Timeout: 60 * time.Second}
	}
}

type HookHTTPClient struct {
	*http.Client
}

func NewHookHTTPClient(timeout SMSHookTimeout) HookHTTPClient {
	return HookHTTPClient{
		utilhttputil.NewExternalClient(timeout.Timeout),
	}
}

type HookDenoClient struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, logger hook.Logger, timeout SMSHookTimeout) HookDenoClient {
	return HookDenoClient{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: utilhttputil.NewExternalClient(timeout.Timeout),
			Logger:     logger,
		},
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

func (c *CustomClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
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
		panic(fmt.Errorf("unsupported hook URL: %v", u))
	}
}
