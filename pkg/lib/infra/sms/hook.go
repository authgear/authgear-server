package sms

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type HookHTTPClient struct {
	*http.Client
}

func NewHookHTTPClient(hookCfg *config.HookConfig, smsCfg *config.CustomSMSProviderConfig) HookHTTPClient {
	if smsCfg != nil && smsCfg.Timeout != nil {
		return HookHTTPClient{
			httputil.NewExternalClient(smsCfg.Timeout.Duration()),
		}
	}
	return HookHTTPClient{
		httputil.NewExternalClient(hookCfg.SyncTimeout.Duration()),
	}
}

type HookDenoClient struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, logger hook.Logger, smsCfg *config.CustomSMSProviderConfig) HookDenoClient {
	var timeout time.Duration
	if smsCfg != nil && smsCfg.Timeout != nil {
		timeout = smsCfg.Timeout.Duration()
	} else {
		timeout = 60 * time.Second
	}

	return HookDenoClient{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(timeout),
			Logger:     logger,
		},
	}
}
