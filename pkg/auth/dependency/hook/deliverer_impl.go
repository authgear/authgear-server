package hook

import (
	"bytes"
	"encoding/json"
	"net"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"

	"github.com/skygeario/skygear-server/pkg/core/hash"
	"github.com/skygeario/skygear-server/pkg/core/http"

	"github.com/franela/goreq"
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
}

func NewDeliverer(config *config.TenantConfiguration, timeProvider time.Provider, mutator Mutator) Deliverer {
	return &delivererImpl{
		Hooks:        &config.Hooks,
		UserConfig:   &config.UserConfig.Hook,
		AppConfig:    &config.AppConfig.Hook,
		TimeProvider: timeProvider,
		Mutator:      mutator,
	}
}

func (deliverer *delivererImpl) DeliverBeforeEvent(e *event.Event, user *model.User) error {
	startTime := deliverer.TimeProvider.Now()
	requestTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTimeout) * gotime.Second
	totalTimeout := gotime.Duration(deliverer.AppConfig.SyncHookTotalTimeout) * gotime.Second

	mutator := deliverer.Mutator.New(e, user)

	for _, hook := range *deliverer.Hooks {
		if hook.Event != string(e.Type) {
			continue
		}

		if deliverer.TimeProvider.Now().Sub(startTime) > totalTimeout {
			return DeliveryTimeout{}
		}

		request, err := deliverer.prepareRequest(e)
		if err != nil {
			return err
		}
		request.Uri = hook.URL
		request.Timeout = requestTimeout

		resp, err := performRequest(request, true)
		if err != nil {
			return err
		}

		if !resp.IsAllowed {
			return OperationDisallowed{
				Items: []OperationDisallowedItem{
					OperationDisallowedItem{
						Reason: resp.Reason,
						Data:   resp.Data,
					},
				},
			}
		}

		if resp.Mutations != nil {
			err = mutator.Add(*resp.Mutations)
			if err != nil {
				return MutationFailed{inner: err}
			}
		}
	}

	err := mutator.Apply()
	if err != nil {
		return MutationFailed{inner: err}
	}

	return nil
}

func (deliverer *delivererImpl) prepareRequest(event *event.Event) (*goreq.Request, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, DeliveryFailed{inner: err}
	}

	signature := hash.HMACSHA256(body, []byte(deliverer.UserConfig.Secret))

	request := goreq.Request{
		Method: "POST",
		Body:   bytes.NewReader(body),
	}
	request.AddHeader("Content-Type", "application/json")
	request.AddHeader(http.HeaderRequestBodySignature, signature)

	return &request, nil
}

func performRequest(request *goreq.Request, withResponse bool) (hookResp *event.HookResponse, err error) {
	var resp *goreq.Response
	resp, err = request.Do()
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		err = DeliveryTimeout{}
		return
	} else if err != nil {
		err = DeliveryFailed{inner: err}
		return
	}

	defer func() {
		closeError := resp.Body.Close()
		if err == nil && closeError != nil {
			err = DeliveryFailed{inner: closeError}
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = DeliveryFailedInvalidStatusCode
		return
	}

	if !withResponse {
		return
	}

	hookResp = &event.HookResponse{}
	err = resp.Body.FromJsonTo(hookResp)
	if err != nil {
		err = DeliveryFailed{inner: err}
		return
	}

	err = hookResp.Validate()
	if err != nil {
		err = DeliveryFailed{inner: err}
		return
	}

	return
}
