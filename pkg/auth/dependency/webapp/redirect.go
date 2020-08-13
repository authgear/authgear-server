package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const DefaultRedirectURI = "/settings"

func GetRedirectURI(r *http.Request, trustProxy bool) string {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		return DefaultRedirectURI
	}
	return redirectURI
}
