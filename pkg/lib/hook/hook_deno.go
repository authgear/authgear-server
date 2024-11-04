package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DenoHookLogger struct{ *log.Logger }

func NewDenoHookLogger(lf *log.Factory) DenoHookLogger { return DenoHookLogger{lf.New("deno-hook")} }

type DenoHook struct {
	ResourceManager ResourceManager
	Logger          DenoHookLogger
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

func (h *DenoHook) RunSync(ctx context.Context, client DenoClient, u *url.URL, input interface{}) (out interface{}, err error) {
	var script []byte
	script, err = h.loadScript(u)
	if err != nil {
		return nil, err
	}

	// Propagate the request context.
	out, err = client.Run(ctx, string(script), input)
	if err != nil {
		return nil, err
	}
	return
}

func (h *DenoHook) RunAsync(ctx context.Context, client DenoClient, u *url.URL, input interface{}) (err error) {
	// Remove cancel from the the context.
	// This is because the hook may finish after the current request finishes.
	ctx = context.WithoutCancel(ctx)

	var script []byte
	script, err = h.loadScript(u)
	if err != nil {
		return
	}

	go func() {
		_, err := client.Run(ctx, string(script), input)
		if err != nil {
			h.Logger.WithError(err).Error("failed to run deno script")
			return
		}
		return
	}()

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

func (h *EventDenoHookImpl) DeliverBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) (*event.HookResponse, error) {
	out, err := h.RunSync(ctx, h.SyncDenoClient, u, e)
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

func (h *EventDenoHookImpl) DeliverNonBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) error {
	return h.RunAsync(ctx, h.AsyncDenoClient, u, e)
}
