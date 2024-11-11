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
	"github.com/authgear/authgear-server/pkg/util/log"
)

type DenoMiddlewareLogger struct{ *log.Logger }

func NewDenoMiddlewareLogger(lf *log.Factory) DenoMiddlewareLogger {
	return DenoMiddlewareLogger{lf.New("account-migration-deno")}
}

type AccountMigrationDenoHook struct {
	hook.DenoHook
	Client HookDenoClient
	Logger DenoMiddlewareLogger
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

	hookResp, err := ParseHookResponse(bytes.NewReader(b))
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = hook.WebHookInvalidResponse.NewWithInfo("invalid response body", apiError.Info)
		return nil, err
	}

	return hookResp, nil
}

var _ Hook = &AccountMigrationDenoHook{}

type HookDenoClient struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, logger hook.Logger, cfg *config.AccountMigrationHookConfig) HookDenoClient {
	return HookDenoClient{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(cfg.Timeout.Duration()),
			Logger:     logger,
		},
	}
}
