package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DenoHookImpl struct {
	Context         context.Context
	SyncDenoClient  SyncDenoClient
	AsyncDenoClient AsyncDenoClient
	ResourceManager ResourceManager
}

var _ DenoHook = &DenoHookImpl{}

func (h *DenoHookImpl) SupportURL(u *url.URL) bool {
	return u.Scheme == "authgeardeno"
}

func (h *DenoHookImpl) DeliverBlockingEvent(u *url.URL, e *event.Event) (*event.HookResponse, error) {
	script, err := h.loadScript(u)
	if err != nil {
		return nil, err
	}

	out, err := h.SyncDenoClient.Run(h.Context, string(script), e)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	hookResp, err := event.ParseHookResponse(bytes.NewReader(b))
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = WebHookInvalidResponse.NewWithInfo("invalid response body", apiError.Info)
		return nil, err
	}

	return hookResp, nil
}

func (h *DenoHookImpl) DeliverNonBlockingEvent(u *url.URL, e *event.Event) error {
	script, err := h.loadScript(u)
	if err != nil {
		return err
	}

	_, err = h.AsyncDenoClient.Run(h.Context, string(script), e)
	if err != nil {
		return err
	}

	return nil
}

func (h *DenoHookImpl) loadScript(u *url.URL) ([]byte, error) {
	out, err := h.ResourceManager.Read(DenoFile, resource.AppFile{
		Path: u.Path,
	})
	if err != nil {
		return nil, err
	}

	return out.([]byte), nil
}
