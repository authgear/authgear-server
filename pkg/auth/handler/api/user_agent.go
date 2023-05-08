package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type WorkflowUserAgentCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

func getOrCreateUserAgentID(cookies WorkflowUserAgentCookieManager, w http.ResponseWriter, r *http.Request) string {
	var userAgentID string
	userAgentIDCookie, err := cookies.GetCookie(r, workflow.UserAgentIDCookieDef)
	if err == nil {
		userAgentID = userAgentIDCookie.Value
	}
	if userAgentID == "" {
		userAgentID = workflow.NewUserAgentID()
	}
	cookie := cookies.ValueCookie(workflow.UserAgentIDCookieDef, userAgentID)
	httputil.UpdateCookie(w, cookie)
	return userAgentID
}
