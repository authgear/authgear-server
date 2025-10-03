package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var WebHookLogger = slogutil.NewLogger("webhook")

type WebHook interface {
	SupportURL(u *url.URL) bool
	PrepareRequest(ctx context.Context, u *url.URL, body interface{}) (*http.Request, error)
	PerformWithResponse(ctx context.Context, client *http.Client, request *http.Request) (resp *http.Response, err error)
	PerformNoResponse(ctx context.Context, client *http.Client, request *http.Request) error
}

type WebHookImpl struct {
	Secret *config.WebhookKeyMaterials
}

var _ WebHook = &WebHookImpl{}

func (h *WebHookImpl) SupportURL(u *url.URL) bool {
	return u.Scheme == "http" || u.Scheme == "https"
}

func (h *WebHookImpl) PrepareRequest(ctx context.Context, u *url.URL, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	key, err := jwkutil.ExtractOctetKey(h.Secret.Set, "")
	if err != nil {
		return nil, err
	}
	signature := crypto.HMACSHA256String(key, jsonBody)

	request, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(HeaderRequestBodySignature, signature)

	return request, nil
}

// The caller should close the response body if the response is not nil.
func (h *WebHookImpl) PerformWithResponse(
	ctx context.Context,
	client *http.Client,
	request *http.Request) (resp *http.Response, err error) {

	resp, err = performRequest(client, request)
	if err != nil {
		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterBlockingWebhookCount,
			otelauthgear.WithStatusError(),
		)
	} else {
		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterBlockingWebhookCount,
			otelauthgear.WithStatusOk(),
		)
	}
	return resp, err
}

func (h *WebHookImpl) PerformNoResponse(
	ctx context.Context,
	client *http.Client,
	request *http.Request) error {
	logger := WebHookLogger.GetLogger(ctx)

	go func() {
		resp, err := performRequest(client, request)
		if err != nil {
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterNonBlockingWebhookCount,
				otelauthgear.WithStatusError(),
			)
			logger.WithError(err).Error(ctx, "failed to dispatch nonblocking webhook")
		} else {
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterNonBlockingWebhookCount,
				otelauthgear.WithStatusOk(),
			)
		}
		if resp != nil {
			defer resp.Body.Close()
		}
	}()

	return nil
}

type EventWebHookImpl struct {
	WebHookImpl
	SyncHTTP  SyncHTTPClient
	AsyncHTTP AsyncHTTPClient
}

var _ EventWebHook = &EventWebHookImpl{}

func (h *EventWebHookImpl) DeliverBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) (*event.HookResponse, error) {
	request, err := h.PrepareRequest(ctx, u, e)
	if err != nil {
		return nil, err
	}

	resp, err := h.PerformWithResponse(ctx, h.SyncHTTP.Client, request)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, err
	}

	var hookResp *event.HookResponse
	hookResp, err = event.ParseHookResponse(ctx, e.Type, resp.Body)
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = HookInvalidResponse.NewWithInfo("invalid response body", apiError.Info_ReadOnly)
		return nil, err
	}

	return hookResp, nil
}

func (h *EventWebHookImpl) DeliverNonBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) error {
	// Detach the deadline so that the context is not canceled along with the request.
	ctx = context.WithoutCancel(ctx)
	request, err := h.PrepareRequest(ctx, u, e)
	if err != nil {
		return err
	}

	return h.PerformNoResponse(ctx, h.AsyncHTTP.Client, request)
}

func performRequest(
	client *http.Client,
	request *http.Request) (resp *http.Response, err error) {
	resp, err = client.Do(request)
	if os.IsTimeout(err) {
		err = HookDeliveryTimeout.New("webhook delivery timeout")
		return
	} else if err != nil {
		err = errors.Join(WebHookDeliveryUnknownFailure.New("failed to deliver webhook"), err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = HookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
	}

	return
}
