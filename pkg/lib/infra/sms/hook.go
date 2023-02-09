package sms

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SMSHookTimeout struct {
	Timeout time.Duration
}

func NewSMSHookTimeout(smsCfg *config.CustomSMSProviderConfig) SMSHookTimeout {
	if smsCfg != nil && smsCfg.Timeout != nil {
		return SMSHookTimeout{Timeout: smsCfg.Timeout.Duration()}
	} else {
		return SMSHookTimeout{Timeout: 60 * time.Second}
	}
}

type HookHTTPClient struct {
	*http.Client
}

func NewHookHTTPClient(timeout SMSHookTimeout) HookHTTPClient {
	return HookHTTPClient{
		httputil.NewExternalClient(timeout.Timeout),
	}
}

type HookDenoClient struct {
	hook.DenoClient
}

func NewHookDenoClient(endpoint config.DenoEndpoint, logger hook.Logger, timeout SMSHookTimeout) HookDenoClient {
	return HookDenoClient{
		&hook.DenoClientImpl{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(timeout.Timeout),
			Logger:     logger,
		},
	}
}
