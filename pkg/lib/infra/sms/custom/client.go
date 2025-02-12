package custom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
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

func NewSMSWebHook(hook hook.WebHook, smsCfg *config.CustomSMSProviderConfig) *SMSWebHook {
	httpClient := NewHookHTTPClient(NewSMSHookTimeout(smsCfg))
	return &SMSWebHook{
		WebHook: hook,
		Client:  httpClient,
	}
}

type SMSWebHook struct {
	hook.WebHook
	Client HookHTTPClient
}

func (w *SMSWebHook) Call(ctx context.Context, u *url.URL, payload SendOptions) error {
	req, err := w.PrepareRequest(ctx, u, payload)
	if err != nil {
		return err
	}

	resp, err := w.Client.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Join(err, &smsapi.SendError{
			DumpedResponse: dumpedResponse,
		})
	}

	responseBody, err := ParseResponseBody(bodyBytes)
	if err != nil {
		// This is not something we understand, check the status code to determine if it is a success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return errors.Join(err, &smsapi.SendError{
			DumpedResponse: dumpedResponse,
		})
	}

	return w.handleResponse(responseBody, dumpedResponse)
}

func (w *SMSWebHook) handleResponse(responseBody *ResponseBody, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse: dumpedResponse,
	}

	errorDetail := func() apierrors.Details {
		d := apierrors.Details{}
		if responseBody.ErrorDetail != nil {
			d["Detail"] = responseBody.ErrorDetail
		}
		return d
	}

	switch responseBody.Code {
	case "ok":
		return nil
	case "invalid_phone_number":
		return errors.Join(smsapi.ErrKindInvalidPhoneNumber.NewWithInfo(
			"phone number rejected by sms gateway", errorDetail()), err)
	case "rate_limited":
		return errors.Join(smsapi.ErrKindRateLimited.NewWithInfo(
			"sms gateway rate limited", errorDetail()), err)
	case "authentication_failed":
		return errors.Join(smsapi.ErrKindRateLimited.NewWithInfo(
			"sms gateway authentication failed", errorDetail()), err)
	case "authorization_failed":
		return errors.Join(smsapi.ErrKindRateLimited.NewWithInfo(
			"sms gateway authorization failed", errorDetail()), err)
	default:
		return err
	}
}

func NewSMSDenoHook(lf *log.Factory, denoEndpoint config.DenoEndpoint, smsCfg *config.CustomSMSProviderConfig) *SMSDenoHook {
	timeout := NewSMSHookTimeout(smsCfg)
	logger := hook.NewLogger(lf)
	client := NewHookDenoClient(denoEndpoint, logger, timeout)
	return &SMSDenoHook{
		Client: client,
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

func (d *SMSDenoHook) Test(ctx context.Context, script string, payload SendOptions) error {
	_, err := d.Client.Run(ctx, script, payload)
	if err != nil {
		return err
	}

	return nil
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
		err = c.SMSWebHook.Call(ctx, u, payload)
		return err
	default:
		panic(fmt.Errorf("unsupported hook URL: %v", u))
	}
}
