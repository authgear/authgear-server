package hook

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

type Deliverer struct {
	Config    *config.HookConfig
	Secret    *config.WebhookKeyMaterials
	Clock     clock.Clock
	SyncHTTP  SyncHTTPClient
	AsyncHTTP AsyncHTTPClient
}

func (deliverer *Deliverer) WillDeliver(eventType event.Type) bool {
	for _, hook := range deliverer.Config.Handlers {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (deliverer *Deliverer) DeliverBeforeEvent(e *event.Event) error {
	startTime := deliverer.Clock.NowMonotonic()
	totalTimeout := deliverer.Config.SyncTotalTimeout.Duration()

	for _, hook := range deliverer.Config.Handlers {
		if hook.Event != string(e.Type) {
			continue
		}

		if deliverer.Clock.NowMonotonic().Sub(startTime) > totalTimeout {
			return errDeliveryTimeout
		}

		request, err := deliverer.prepareRequest(hook, e)
		if err != nil {
			return err
		}

		resp, err := performRequest(deliverer.SyncHTTP.Client, request, true)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return newErrorOperationDisallowed(
				[]OperationDisallowedItem{{
					Reason: resp.Reason,
					Data:   resp.Data,
				}},
			)
		}
	}

	return nil
}

func (deliverer *Deliverer) DeliverNonBeforeEvent(e *event.Event) error {
	for _, hook := range deliverer.Config.Handlers {
		if hook.Event != string(e.Type) {
			continue
		}

		request, err := deliverer.prepareRequest(hook, e)
		if err != nil {
			return err
		}

		_, err = performRequest(deliverer.AsyncHTTP.Client, request, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (deliverer *Deliverer) prepareRequest(hook config.HookHandlerConfig, event *event.Event) (*http.Request, error) {
	hookURL, err := url.Parse(hook.URL)
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}

	key, err := jwkutil.ExtractOctetKey(&deliverer.Secret.Set, "")
	if err != nil {
		panic("hook: web-hook key not found")
	}
	signature := crypto.HMACSHA256String(key, body)

	request, err := http.NewRequest("POST", hookURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(HeaderRequestBodySignature, signature)

	return request, nil
}

func performRequest(client *http.Client, request *http.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *http.Response
	resp, err = client.Do(request)
	if reqError, ok := err.(net.Error); ok && reqError.Timeout() {
		err = errDeliveryTimeout
		return
	} else if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	defer func() {
		closeError := resp.Body.Close()
		if err == nil && closeError != nil {
			err = newErrorDeliveryFailed(closeError)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errDeliveryInvalidStatusCode
		return
	}

	if !withResponse {
		return
	}

	hookResp, err = event.ParseHookResponse(resp.Body)
	if err != nil {
		err = newErrorDeliveryFailed(err)
		return
	}

	return
}
