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

func NewHookDenoClient(endpoint config.DenoEndpoint, timeout SMSHookTimeout) HookDenoClient {
	return HookDenoClientImpl{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: utilhttputil.NewExternalClient(timeout.Timeout),
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

	return handleResponse("webhook", responseBody, dumpedResponse)
}

func NewSMSDenoHookForTest(denoEndpoint config.DenoEndpoint, smsCfg *config.CustomSMSProviderConfig) *SMSDenoHook {
	timeout := NewSMSHookTimeout(smsCfg)
	client := NewHookDenoClient(denoEndpoint, timeout)
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

	return handleResponse("denohook", responseBody, jsonText)
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

func handleResponse(gatewayType string, responseBody *ResponseBody, dumpedResponse []byte) error {

	err := &smsapi.SendError{
		DumpedResponse: dumpedResponse,
	}
	if responseBody.ProviderName != "" {
		err.ProviderName = responseBody.ProviderName
	} else {
		err.ProviderName = gatewayType
	}
	if responseBody.ProviderType != "" {
		err.ProviderType = responseBody.ProviderType
	}
	if responseBody.ProviderErrorCode != "" {
		err.ProviderErrorCode = responseBody.ProviderErrorCode
	}

	switch responseBody.Code {
	case "ok":
		return nil
	case "invalid_phone_number":
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	case "rate_limited":
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case "authentication_failed":
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case "delivery_rejected":
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	case "attempted_to_send_otp_template_without_code":
		err.APIErrorKind = &smsapi.ErrKindAttemptedToSendOTPTemplateWithoutCode
	}
	return err
}
