package hook

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const HeaderRequestBodySignature = "x-authgear-body-signature"

type SyncHTTPClient struct {
	*http.Client
}

func NewSyncHTTPClient(c *config.HookConfig, f *config.HTTPFeatureConfig) SyncHTTPClient {
	return SyncHTTPClient{
		httputil.NewSSRFSafeExternalClient(c.SyncTimeout.Duration(), httputil.SSRFSafeExternalClientOptions{
			AllowNonPublicAddresses: f.IsInsecureFetchAddressAllowed(),
			AllowedHosts:            f.GetInsecureFetchAddressAllowedHosts(),
			Sink:                    "hook.blocking_handlers",
		}),
	}
}

type AsyncHTTPClient struct {
	*http.Client
}

func NewAsyncHTTPClient(f *config.HTTPFeatureConfig) AsyncHTTPClient {
	return AsyncHTTPClient{
		httputil.NewSSRFSafeExternalClient(60*time.Second, httputil.SSRFSafeExternalClientOptions{
			AllowNonPublicAddresses: f.IsInsecureFetchAddressAllowed(),
			AllowedHosts:            f.GetInsecureFetchAddressAllowedHosts(),
			Sink:                    "hook.non_blocking_handlers",
		}),
	}
}
