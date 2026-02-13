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
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var DenoHookLogger = slogutil.NewLogger("deno-hook")

type DenoHook struct {
	ResourceManager ResourceManager
}

func (h *DenoHook) SupportURL(u *url.URL) bool {
	return u.Scheme == "authgeardeno"
}

func (h *DenoHook) loadScript(ctx context.Context, u *url.URL) ([]byte, error) {
	out, err := h.ResourceManager.Read(ctx, DenoFile, resource.AppFile{
		Path: h.rel(u.Path),
	})
	if err != nil {
		return nil, err
	}

	return out.([]byte), nil
}

func (h *DenoHook) RunSync(ctx context.Context, client DenoClient, u *url.URL, input interface{}) (out interface{}, err error) {
	var script []byte
	script, err = h.loadScript(ctx, u)
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
	logger := DenoHookLogger.GetLogger(ctx)
	// Remove cancel from the the context.
	// This is because the hook may finish after the current request finishes.
	ctx = context.WithoutCancel(ctx)

	var script []byte
	script, err = h.loadScript(ctx, u)
	if err != nil {
		return
	}

	go func() {
		_, err := client.Run(ctx, string(script), input)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to run deno script")
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

	hookResp, err := event.ParseHookResponse(ctx, e.Type, bytes.NewReader(b))
	if err != nil {
		apiError := apierrors.AsAPIErrorWithContext(ctx, err)
		err = HookInvalidResponse.NewWithInfo("invalid response body", apiError.Info_ReadOnly)
		return nil, err
	}

	return hookResp, nil
}

func (h *EventDenoHookImpl) DeliverNonBlockingEvent(ctx context.Context, u *url.URL, e *event.Event) error {
	return h.RunAsync(ctx, h.AsyncDenoClient, u, e)
}
