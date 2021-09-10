package authenticationinfo

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var CookieDef = &httputil.CookieDef{
	NameSuffix: "authentication_info",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}
