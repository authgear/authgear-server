package workflow

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieGetter interface {
	GetCookies(ctx context.Context, deps *Dependencies, workflows Workflows) ([]*http.Cookie, error)
}

func NewUserAgentIDCookieDef() *httputil.CookieDef {
	def := &httputil.CookieDef{
		NameSuffix:        "workflow_ua_id",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode,
	}
	return def
}

var UserAgentIDCookieDef = NewUserAgentIDCookieDef()
