package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
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

func (w *SMSWebHook) Call(ctx context.Context, u *url.URL, payload SendOptions) ([]byte, error) {
	req, err := w.PrepareRequest(ctx, u, payload)
	if err != nil {
		return nil, err
	}

	resp, err := w.Client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return dumpedResponse, nil
	}

	return nil, &smsapi.SendError{
		DumpedResponse: dumpedResponse,
	}
}

type SMSDenoHook struct {
	hook.DenoHook
	Client HookDenoClient
}

func (d *SMSDenoHook) Call(ctx context.Context, u *url.URL, payload SendOptions) ([]byte, error) {
	anything, err := d.RunSync(ctx, d.Client, u, payload)
	if err != nil {
		return nil, err
	}

	jsonText, err := json.Marshal(anything)
	if err != nil {
		return nil, err
	}

	return jsonText, nil
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
	payload := SendOptions{
		To:                opts.To,
		Body:              opts.Body,
		AppID:             opts.AppID,
		TemplateName:      opts.TemplateName,
		LanguageTag:       opts.LanguageTag,
		TemplateVariables: opts.TemplateVariables,
	}
	switch {
	case c.SMSDenoHook.SupportURL(u):
		_, err = c.SMSDenoHook.Call(ctx, u, payload)
		return err
	case c.SMSWebHook.SupportURL(u):
		_, err = c.SMSWebHook.Call(ctx, u, payload)
		return err
	default:
		panic(fmt.Errorf("unsupported hook URL: %v", u))
	}
}
