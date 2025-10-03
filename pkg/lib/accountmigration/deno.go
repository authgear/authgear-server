package accountmigration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AccountMigrationDenoHook struct {
	hook.DenoHook
	Client HookDenoClient
}

func (h *AccountMigrationDenoHook) Call(ctx context.Context, u *url.URL, hookReq *HookRequest) (*HookResponse, error) {
	out, err := h.RunSync(ctx, h.Client, u, hookReq)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	hookResp, err := ParseHookResponse(ctx, bytes.NewReader(b))
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = hook.HookInvalidResponse.NewWithInfo("invalid response body", apiError.Info_ReadOnly)
		return nil, err
	}

	return hookResp, nil
}

var _ Hook = &AccountMigrationDenoHook{}

type HookDenoClient struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, cfg *config.AccountMigrationHookConfig) HookDenoClient {
	return HookDenoClient{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(cfg.Timeout.Duration()),
		},
	}
}
