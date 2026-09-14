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
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

// DefaultSMSHookTimeout is what both timeout constructors fall back to when
// the deployment or the project did not configure one.
const DefaultSMSHookTimeout = 60 * time.Second

type SMSHookTimeout struct {
	Timeout time.Duration
}

func NewSMSHookTimeout(smsCfg *config.CustomSMSProviderConfig) SMSHookTimeout {
	if smsCfg != nil && smsCfg.Timeout != nil {
		return SMSHookTimeout{Timeout: smsCfg.Timeout.Duration()}
	} else {
		return SMSHookTimeout{Timeout: DefaultSMSHookTimeout}
	}
}

// EnvSMSHookTimeout carries SMS_GATEWAY_CUSTOM_TIMEOUT, which is a number of
// seconds. It is kept apart from SMSHookTimeout because the two come from
// different sources and reach different clients.
type EnvSMSHookTimeout struct {
	Timeout time.Duration
}

// NewEnvSMSHookTimeout never returns a zero duration: zero on
// http.Client.Timeout means no timeout at all, which is not what an unset or
// malformed SMS_GATEWAY_CUSTOM_TIMEOUT asks for.
func NewEnvSMSHookTimeout(envCfg config.SMSGatewayEnvironmentCustomSMSProviderConfig) EnvSMSHookTimeout {
	seconds, err := strconv.Atoi(envCfg.Timeout)
	if err != nil || seconds <= 0 {
		return EnvSMSHookTimeout{Timeout: DefaultSMSHookTimeout}
	}
	return EnvSMSHookTimeout{Timeout: time.Duration(seconds) * time.Second}
}

type HookHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HookHTTPClientImpl struct {
	*http.Client
}

// NewHookHTTPClient is for the custom SMS provider a project configures in
// authgear.secrets.yaml. A project admin chooses that URL, so it is bound by
// the fetch address policy; see docs/specs/ssrf-protection.md.
func NewHookHTTPClient(timeout SMSHookTimeout, f *config.HTTPFeatureConfig) HookHTTPClient {
	return HookHTTPClientImpl{
		utilhttputil.NewSSRFSafeExternalClient(timeout.Timeout, utilhttputil.SSRFSafeExternalClientOptions{
			AllowNonPublicAddresses: f.IsInsecureFetchAddressAllowed(),
			AllowedHosts:            f.GetInsecureFetchAddressAllowedHosts(),
			Sink:                    "sms.custom.url",
		}),
	}
}

// EnvHookHTTPClient is a distinct interface from HookHTTPClient, not an alias
// of it: wire has to be able to provide both in one injector, and nothing that
// fetches a project-supplied URL should be able to pick this one up by
// accident.
type EnvHookHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type EnvHookHTTPClientImpl struct {
	*http.Client
}

// NewEnvHookHTTPClient is for the SMS gateway this deployment configures
// through SMS_GATEWAY_CUSTOM_URL. It deliberately applies no address policy:
// the operator chose the destination, and running authgear-sms-gateway beside
// Authgear in the same cluster -- so on an address that is not publicly
// routable -- is the normal deployment. Same reasoning as the Deno hook runner
// and object storage; see docs/specs/ssrf-protection.md.
func NewEnvHookHTTPClient(timeout EnvSMSHookTimeout) EnvHookHTTPClient {
	return EnvHookHTTPClientImpl{
		utilhttputil.NewExternalClient(timeout.Timeout),
	}
}

type HookDenoClient interface {
	Run(ctx context.Context, script string, input any) (out any, err error)
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

func NewSMSWebHook(hook hook.WebHook, smsCfg *config.CustomSMSProviderConfig, f *config.HTTPFeatureConfig) *SMSWebHook {
	httpClient := NewHookHTTPClient(NewSMSHookTimeout(smsCfg), f)
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
	return callWebHook(ctx, w.WebHook, w.Client, u, payload)
}

// EnvSMSWebHook calls the SMS gateway named by SMS_GATEWAY_CUSTOM_URL. It
// differs from SMSWebHook only in which client it holds; the request and the
// response handling are the same, which is why both delegate to callWebHook.
type EnvSMSWebHook struct {
	hook.WebHook
	Client EnvHookHTTPClient
}

func (w *EnvSMSWebHook) Call(ctx context.Context, u *url.URL, payload SendOptions) error {
	return callWebHook(ctx, w.WebHook, w.Client, u, payload)
}

type webHookDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func callWebHook(ctx context.Context, wh hook.WebHook, client webHookDoer, u *url.URL, payload SendOptions) error {
	req, err := wh.PrepareRequest(ctx, u, payload)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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
	RunSync(ctx context.Context, client hook.DenoClient, u *url.URL, input any) (out any, err error)
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

func (d *SMSDenoHook) handleOutput(output any) error {
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

// SMSWebHookCaller is what CustomClient needs of a webhook, so that the same
// client serves both the project-configured URL and the one this deployment
// configures.
type SMSWebHookCaller interface {
	SupportURL(u *url.URL) bool
	Call(ctx context.Context, u *url.URL, payload SendOptions) error
}

var _ SMSWebHookCaller = (*SMSWebHook)(nil)
var _ SMSWebHookCaller = (*EnvSMSWebHook)(nil)

type CustomClient struct {
	Config      *config.CustomSMSProviderConfig
	SMSDenoHook SMSDenoHook
	SMSWebHook  SMSWebHookCaller
}

var _ smsapi.Client = (*CustomClient)(nil)

func NewCustomClient(c *config.CustomSMSProviderConfig, d SMSDenoHook, w SMSWebHookCaller) *CustomClient {
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
		ProviderType:   config.SMSProviderCustom,
		DumpedResponse: dumpedResponse,
	}

	// Handle info returned by authgear-sms-gateway
	if responseBody.Info != nil {
		if providerName, ok := responseBody.Info["provider_name"].(string); ok {
			err.CustomProviderName = providerName
		}
		if providerType, ok := responseBody.Info["provider_type"].(string); ok {
			err.CustomProviderType = providerType
		}
		if providerErrorCode, ok := responseBody.Info["provider_error_code"].(string); ok {
			err.ProviderErrorCode = providerErrorCode
		}
	}

	err.CustomProviderResponseCode = responseBody.Code
	err.CustomProviderDescription = responseBody.Description
	switch responseBody.Code {
	case "ok":
		return nil
	case "invalid_phone_number":
		err.APIErrorKind = &smsapi.ErrKindInvalidPhoneNumber
	case "rate_limited":
		err.APIErrorKind = &smsapi.ErrKindRateLimited
	case "unsupported_request":
		err.APIErrorKind = &smsapi.ErrKindUnsupportedRequest
	case "authentication_failed":
		err.APIErrorKind = &smsapi.ErrKindAuthenticationFailed
	case "delivery_rejected":
		err.APIErrorKind = &smsapi.ErrKindDeliveryRejected
	case "timeout":
		err.APIErrorKind = &smsapi.ErrKindTimeout
	}
	return err
}
