package hook

import (
	"bytes"
	"encoding/json"
	"net"
	gohttp "net/http"
	"net/url"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/time"

	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/http"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type delivererImpl struct {
	Hooks        *[]config.Hook
	UserConfig   *config.HookUserConfiguration
	AppConfig    *config.HookAppConfiguration
	TimeProvider time.Provider
	Mutator      Mutator
	HTTPClient   gohttp.Client
}

func NewDeliverer(config *config.TenantConfiguration, timeProvider time.Provider, mutator Mutator) Deliverer {
	return &delivererImpl{
		Hooks:        &config.Hooks,
		UserConfig:   config.UserConfig.Hook,
		AppConfig:    &config.AppConfig.Hook,
		TimeProvider: timeProvider,
		Mutator:      mutator,
		HTTPClient:   gohttp.Client{},
	}
}

func (deliverer *delivererImpl) WillDeliver(eventType event.Type) bool {
	for _, hook := range *deliverer.Hooks {
		if hook.Event == string(eventType) {
			return true
		}
	}
	return false
}

func (deliverer *delivererImpl) DeliverBeforeEvent(baseURL *url.URL, e *event.Event, user *model.User) error {
	startTime := deliverer.TimeProvider.Now()
	requestTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTimeout) * gotime.Second
	totalTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTotalTimeout) * gotime.Second

	mutator := deliverer.Mutator.New(e, user)
	client := deliverer.HTTPClient
	client.CheckRedirect = noFollowRedirectPolicy
	client.Timeout = requestTimeout

	for _, hook := range *deliverer.Hooks {
		if hook.Event != string(e.Type) {
			continue
		}

		if deliverer.TimeProvider.Now().Sub(startTime) > totalTimeout {
			return errDeliveryTimeout
		}

		request, err := deliverer.prepareRequest(baseURL, hook, e)
		if err != nil {
			return err
		}

		resp, err := performRequest(client, request, true)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return newErrorOperationDisallowed(
				[]OperationDisallowedItem{
					OperationDisallowedItem{
						Reason: resp.Reason,
						Data:   resp.Data,
					},
				},
			)
		}

		if resp.Mutations != nil {
			err = mutator.Add(*resp.Mutations)
			if err != nil {
				return newErrorMutationFailed(err)
			}
		}
	}

	err := mutator.Apply()
	if err != nil {
		return newErrorMutationFailed(err)
	}

	return nil
}

func (deliverer *delivererImpl) DeliverNonBeforeEvent(baseURL *url.URL, e *event.Event, timeout gotime.Duration) error {
	client := deliverer.HTTPClient
	client.CheckRedirect = noFollowRedirectPolicy
	client.Timeout = timeout

	for _, hook := range *deliverer.Hooks {
		if hook.Event != string(e.Type) {
			continue
		}

		request, err := deliverer.prepareRequest(baseURL, hook, e)
		if err != nil {
			return err
		}

		_, err = performRequest(client, request, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (deliverer *delivererImpl) prepareRequest(baseURL *url.URL, hook config.Hook, event *event.Event) (*gohttp.Request, error) {
	hookURL, err := url.Parse(hook.URL)
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}
	hookURL = baseURL.ResolveReference(hookURL)

	body, err := json.Marshal(event)
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}

	signature := crypto.HMACSHA256String([]byte(deliverer.UserConfig.Secret), body)

	request, err := gohttp.NewRequest("POST", hookURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, newErrorDeliveryFailed(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(http.HeaderRequestBodySignature, signature)

	return request, nil
}

func noFollowRedirectPolicy(*gohttp.Request, []*gohttp.Request) error {
	return gohttp.ErrUseLastResponse
}

func performRequest(client gohttp.Client, request *gohttp.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *gohttp.Response
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
