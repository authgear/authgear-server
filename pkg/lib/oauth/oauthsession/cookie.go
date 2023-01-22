package oauthsession

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// CookieDef contains oauth_session.
// It is private to oauth.
var CookieDef = &httputil.CookieDef{
	NameSuffix: "oauth_session",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// UICookieDef contains oauth_ui.
// It is supposed to be set and pop immediately to create a ui session.
var UICookieDef = &httputil.CookieDef{
	NameSuffix: "oauth_ui",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}
