package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

//go:generate mockgen -source=deliverer.go -destination=deliverer_mock_test.go -package hook

type StdAttrsServiceNoEvent interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type Deliverer struct {
	Config                 *config.HookConfig
	Secret                 *config.WebhookKeyMaterials
	Clock                  clock.Clock
	SyncHTTP               SyncHTTPClient
	AsyncHTTP              AsyncHTTPClient
	StdAttrsServiceNoEvent StdAttrsServiceNoEvent
}

func (deliverer *Deliverer) DeliverBlockingEvent(e *event.Event) error {
	startTime := deliverer.Clock.NowMonotonic()
	totalTimeout := deliverer.Config.SyncTotalTimeout.Duration()

	mutationsEverApplied := false
	for _, hook := range deliverer.Config.BlockingHandlers {
		if hook.Event != string(e.Type) {
			continue
		}

		elapsed := deliverer.Clock.NowMonotonic().Sub(startTime)
		if elapsed > totalTimeout {
			return WebHookDeliveryTimeout.NewWithInfo("webhook delivery timeout", apierrors.Details{
				"elapsed": elapsed,
				"limit":   totalTimeout,
			})
		}

		request, err := deliverer.prepareRequest(hook.URL, e)
		if err != nil {
			return err
		}

		resp, err := performRequest(deliverer.SyncHTTP.Client, request, true)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return newErrorOperationDisallowed(
				string(e.Type),
				[]OperationDisallowedItem{{
					Title:  resp.Title,
					Reason: resp.Reason,
				}},
			)
		}

		var applied bool
		e, applied = e.ApplyMutations(resp.Mutations)
		if applied {
			mutationsEverApplied = true
		}
	}

	if mutationsEverApplied {
		if mutations, ok := e.GenerateFullMutations(); ok {
			if mutations.User.StandardAttributes != nil {
				userID := e.Payload.UserID()
				err := deliverer.StdAttrsServiceNoEvent.UpdateStandardAttributes(
					config.RolePortalUI,
					userID,
					mutations.User.StandardAttributes,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (deliverer *Deliverer) DeliverNonBlockingEvent(e *event.Event) error {
	if !e.IsNonBlocking {
		return nil
	}

	checkDeliver := func(events []string, target string) bool {
		for _, event := range events {
			if event == "*" {
				return true
			}
			if event == target {
				return true
			}
		}
		return false
	}

	for _, hook := range deliverer.Config.NonBlockingHandlers {
		shouldDeliver := checkDeliver(hook.Events, string(e.Type))
		if !shouldDeliver {
			continue
		}

		request, err := deliverer.prepareRequest(hook.URL, e)
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

func (deliverer *Deliverer) WillDeliverBlockingEvent(eventType event.Type) bool {
	for _, hook := range deliverer.Config.BlockingHandlers {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (deliverer *Deliverer) WillDeliverNonBlockingEvent(eventType event.Type) bool {
	for _, hook := range deliverer.Config.NonBlockingHandlers {
		for _, e := range hook.Events {
			if e == "*" {
				return true
			}
			if e == string(eventType) {
				return true
			}
		}
	}
	return false
}

func (deliverer *Deliverer) prepareRequest(urlStr string, event *event.Event) (*http.Request, error) {
	hookURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}

	key, err := jwkutil.ExtractOctetKey(deliverer.Secret.Set, "")
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}
	signature := crypto.HMACSHA256String(key, body)

	request, err := http.NewRequest("POST", hookURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(HeaderRequestBodySignature, signature)

	return request, nil
}

func performRequest(client *http.Client, request *http.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *http.Response
	resp, err = client.Do(request)
	if reqError, ok := err.(net.Error); ok && reqError.Timeout() {
		err = WebHookDeliveryTimeout.New("webhook delivery timeout")
		return
	} else if err != nil {
		err = fmt.Errorf("webhook: %w", err)
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
