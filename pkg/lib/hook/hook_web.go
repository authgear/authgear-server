package hook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

type WebHookImpl struct {
	Secret    *config.WebhookKeyMaterials
	SyncHTTP  SyncHTTPClient
	AsyncHTTP AsyncHTTPClient
}

var _ WebHook = &WebHookImpl{}

func (h *WebHookImpl) DeliverBlockingEvent(cfg config.BlockingHandlersConfig, e *event.Event) (*event.HookResponse, error) {
	request, err := h.prepareRequest(cfg.URL, e)
	if err != nil {
		return nil, err
	}

	resp, err := h.performRequest(h.SyncHTTP.Client, request, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *WebHookImpl) DeliverNonBlockingEvent(cfg config.NonBlockingHandlersConfig, e *event.Event) error {
	request, err := h.prepareRequest(cfg.URL, e)
	if err != nil {
		return err
	}

	_, err = h.performRequest(h.AsyncHTTP.Client, request, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *WebHookImpl) prepareRequest(urlStr string, event *event.Event) (*http.Request, error) {
	hookURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	key, err := jwkutil.ExtractOctetKey(h.Secret.Set, "")
	if err != nil {
		return nil, err
	}
	signature := crypto.HMACSHA256String(key, body)

	request, err := http.NewRequest("POST", hookURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(HeaderRequestBodySignature, signature)

	return request, nil
}

func (h *WebHookImpl) performRequest(client *http.Client, request *http.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *http.Response
	resp, err = client.Do(request)
	if os.IsTimeout(err) {
		err = WebHookDeliveryTimeout.New("webhook delivery timeout")
		return
	} else if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = WebHookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
		return
	}

	if !withResponse {
		return
	}

	hookResp, err = event.ParseHookResponse(resp.Body)
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = WebHookInvalidResponse.NewWithInfo("invalid response body", apiError.Info)
		return
	}

	return
}
