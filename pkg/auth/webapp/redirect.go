package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func GetRedirectURI(r *http.Request, trustProxy bool, defaultURI string) string {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		return defaultURI
	}
	return redirectURI
}

func DefaultRedirectURI(uiConfig *config.UIConfig) string {
	if uiConfig != nil && uiConfig.HomeURI != "" {
		return uiConfig.HomeURI
	}
	return "/settings"
}
