package workflow

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieGetter interface {
	GetCookies(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]*http.Cookie, error)
}

func NewUserAgentIDCookieDef() *httputil.CookieDef {
	maxAge := int(Lifetime.Seconds())
	def := &httputil.CookieDef{
		NameSuffix:        "workflow_ua_id",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode,
		MaxAge:            &maxAge,
	}
	return def
}

var UserAgentIDCookieDef = NewUserAgentIDCookieDef()
