package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DenoHook struct {
	Context         context.Context
	ResourceManager ResourceManager
}

func (h *DenoHook) SupportURL(u *url.URL) bool {
	return u.Scheme == "authgeardeno"
}

func (h *DenoHook) loadScript(u *url.URL) ([]byte, error) {
	out, err := h.ResourceManager.Read(DenoFile, resource.AppFile{
		Path: h.rel(u.Path),
	})
	if err != nil {
		return nil, err
	}

	return out.([]byte), nil
}

func (h *DenoHook) Run(client DenoClient, u *url.URL, input interface{}) (out interface{}, err error) {
	var script []byte
	script, err = h.loadScript(u)
	if err != nil {
		return nil, err
	}

	out, err = client.Run(h.Context, string(script), input)
	if err != nil {
		return nil, err
	}
	return
}

// rel is a simplified version of filepath.Rel.
func (h *DenoHook) rel(p string) string {
	return strings.TrimPrefix(p, "/")
}

type EventDenoHookImpl struct {
	DenoHook
	SyncDenoClient  SyncDenoClient
	AsyncDenoClient AsyncDenoClient
}

var _ EventDenoHook = &EventDenoHookImpl{}

func (h *EventDenoHookImpl) DeliverBlockingEvent(u *url.URL, e *event.Event) (*event.HookResponse, error) {
	out, err := h.Run(h.SyncDenoClient, u, e)
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

func (h *EventDenoHookImpl) DeliverNonBlockingEvent(u *url.URL, e *event.Event) error {
	_, err := h.Run(h.AsyncDenoClient, u, e)
	return err
}
