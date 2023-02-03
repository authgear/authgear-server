package sms

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type HookHTTPClient struct {
	*http.Client
}

func NewHookHTTPClient(hookCfg *config.HookConfig, smsCfg *config.CustomSMSProviderConfig) HookHTTPClient {
	if smsCfg != nil {
		return HookHTTPClient{
			httputil.NewExternalClient(smsCfg.Timeout.Duration()),
		}
	}
	return HookHTTPClient{
		httputil.NewExternalClient(hookCfg.SyncTimeout.Duration()),
	}
}
