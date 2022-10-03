package oauthsession

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var CookieDef = &httputil.CookieDef{
	NameSuffix: "oauth_session",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}
