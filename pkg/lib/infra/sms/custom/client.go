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

type HookHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HookHTTPClientImpl struct {
	*http.Client
}

func NewHookHTTPClient(timeout SMSHookTimeout) HookHTTPClient {
	return HookHTTPClientImpl{
		utilhttputil.NewExternalClient(timeout.Timeout),
	}
}

type HookDenoClient interface {
	Run(ctx context.Context, script string, input interface{}) (out interface{}, err error)
}

type HookDenoClientImpl struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, logger hook.Logger, timeout SMSHookTimeout) HookDenoClient {
	return HookDenoClientImpl{
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

	resp, err := w.Client.Do(req)
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
		// Ignore the parse error, return smsapi.SendError with dumped response
		return &smsapi.SendError{
			DumpedResponse: dumpedResponse,
		}
	}

	return handleResponse(responseBody, dumpedResponse)
}

func NewSMSDenoHookForTest(lf *log.Factory, denoEndpoint config.DenoEndpoint, smsCfg *config.CustomSMSProviderConfig) *SMSDenoHook {
	timeout := NewSMSHookTimeout(smsCfg)
	logger := hook.NewLogger(lf)
	client := NewHookDenoClient(denoEndpoint, logger, timeout)
	// DenoHook is not needed because it can only be used for Test()
	return &SMSDenoHook{
		Client: client,
	}
}

type DenoHook interface {
	RunSync(ctx context.Context, client hook.DenoClient, u *url.URL, input interface{}) (out interface{}, err error)
	SupportURL(u *url.URL) bool
}

type SMSDenoHook struct {
	DenoHook
	Client HookDenoClient
}

func (d *SMSDenoHook) Call(ctx context.Context, u *url.URL, payload SendOptions) error {
	anything, err := d.RunSync(ctx, d.Client, u, payload)
	if err != nil {
		return err
	}

	return d.handleOutput(anything)
}

func (d *SMSDenoHook) Test(ctx context.Context, script string, payload SendOptions) error {
	anything, err := d.Client.Run(ctx, script, payload)
	if err != nil {
		return err
	}

	return d.handleOutput(anything)
}

func (d *SMSDenoHook) handleOutput(output interface{}) error {
	if output == nil {
		// This is a null, but we should still consider it is a success for backward compatibility.
		return nil
	}

	jsonText, err := json.Marshal(output)
	if err != nil {
		return err
	}

	responseBody, err := ParseResponseBody(jsonText)
	if err != nil {
		// This is not something we understand, but still consider it is a success for backward compatibility.
		var jsonErr *json.UnmarshalTypeError
		if errors.As(err, &jsonErr) {
			return nil
		}
		return errors.Join(err, &smsapi.SendError{
			DumpedResponse: jsonText,
		})
	}

	return handleResponse(responseBody, jsonText)
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
		err = c.SMSDenoHook.Call(ctx, u, payload)
		return err
	case c.SMSWebHook.SupportURL(u):
		err = c.SMSWebHook.Call(ctx, u, payload)
		return err
	default:
		panic(fmt.Errorf("unsupported hook URL: %v", u))
	}
}

func handleResponse(responseBody *ResponseBody, dumpedResponse []byte) error {
	err := &smsapi.SendError{
		DumpedResponse: dumpedResponse,
	}

	errorDetail := func() apierrors.Details {
		d := apierrors.Details{}
		if responseBody.ProviderName != "" {
			d["ProviderName"] = responseBody.ProviderName
		} else {
			d["ProviderName"] = "webhook"
		}
		if responseBody.ProviderErrorCode != "" {
			d["ProviderErrorCode"] = responseBody.ProviderErrorCode
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
		return errors.Join(smsapi.ErrKindAuthenticationFailed.NewWithInfo(
			"sms gateway authentication failed", errorDetail()), err)
	case "delivery_rejected":
		return errors.Join(smsapi.ErrKindDeliveryRejected.NewWithInfo(
			"sms gateway delievery rejected", errorDetail()), err)
	default:
		return err
	}
}
