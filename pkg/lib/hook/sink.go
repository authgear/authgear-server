package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=sink.go -destination=sink_mock_test.go -package hook

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook-sink")} }

type StandardAttributesServiceNoEvent interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesServiceNoEvent interface {
	UpdateAllCustomAttributes(role accesscontrol.Role, userID string, reprForm map[string]interface{}) error
}

type Sink struct {
	Logger             Logger
	Config             *config.HookConfig
	Secret             *config.WebhookKeyMaterials
	Clock              clock.Clock
	SyncHTTP           SyncHTTPClient
	AsyncHTTP          AsyncHTTPClient
	StandardAttributes StandardAttributesServiceNoEvent
	CustomAttributes   CustomAttributesServiceNoEvent
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
	if s.WillDeliverBlockingEvent(e.Type) {
		err = s.DeliverBlockingEvent(e)
		if err != nil {
			if !apierrors.IsKind(err, WebHookDisallowed) {
				err = fmt.Errorf("failed to dispatch event: %w", err)
			}
			return
		}
	}

	return
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) (err error) {
	// Skip events that are not for webhook.
	payload := e.Payload.(event.NonBlockingPayload)
	if !payload.ForHook() {
		return
	}

	if s.WillDeliverNonBlockingEvent(e.Type) {
		if err := s.DeliverNonBlockingEvent(e); err != nil {
			s.Logger.WithError(err).Error("failed to dispatch non blocking event")
		}
	}

	return
}

func (s *Sink) DeliverBlockingEvent(e *event.Event) error {
	startTime := s.Clock.NowMonotonic()
	totalTimeout := s.Config.SyncTotalTimeout.Duration()

	mutationsEverApplied := false
	for _, hook := range s.Config.BlockingHandlers {
		if hook.Event != string(e.Type) {
			continue
		}

		elapsed := s.Clock.NowMonotonic().Sub(startTime)
		if elapsed > totalTimeout {
			return WebHookDeliveryTimeout.NewWithInfo("webhook delivery timeout", apierrors.Details{
				"elapsed": elapsed,
				"limit":   totalTimeout,
			})
		}

		request, err := s.prepareRequest(hook.URL, e)
		if err != nil {
			return err
		}

		resp, err := performRequest(s.SyncHTTP.Client, request, true)
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
		userID := e.Payload.UserID()
		if mutations, ok := e.GenerateFullMutations(); ok {
			if mutations.User.StandardAttributes != nil {
				err := s.StandardAttributes.UpdateStandardAttributes(
					accesscontrol.RoleGreatest,
					userID,
					mutations.User.StandardAttributes,
				)
				if err != nil {
					return err
				}
			}
			if mutations.User.CustomAttributes != nil {
				err := s.CustomAttributes.UpdateAllCustomAttributes(
					accesscontrol.RoleGreatest,
					userID,
					mutations.User.CustomAttributes,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Sink) DeliverNonBlockingEvent(e *event.Event) error {
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

	for _, hook := range s.Config.NonBlockingHandlers {
		shouldDeliver := checkDeliver(hook.Events, string(e.Type))
		if !shouldDeliver {
			continue
		}

		request, err := s.prepareRequest(hook.URL, e)
		if err != nil {
			return err
		}

		_, err = performRequest(s.AsyncHTTP.Client, request, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sink) WillDeliverBlockingEvent(eventType event.Type) bool {
	for _, hook := range s.Config.BlockingHandlers {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (s *Sink) WillDeliverNonBlockingEvent(eventType event.Type) bool {
	for _, hook := range s.Config.NonBlockingHandlers {
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

func (s *Sink) prepareRequest(urlStr string, event *event.Event) (*http.Request, error) {
	hookURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("webhook: %w", err)
	}

	key, err := jwkutil.ExtractOctetKey(s.Secret.Set, "")
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
	if os.IsTimeout(err) {
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
