package oauthsession

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// UICookieDef contains oauth_ui.
// It is supposed to be set and pop immediately to create a ui session.
// UICookieDef is deprecated, and will be removed.
var UICookieDef = &httputil.CookieDef{
	NameSuffix: "oauth_ui",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}
