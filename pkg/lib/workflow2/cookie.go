package workflow2

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
		NameSuffix:        "workflow2_ua_id",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode,
	}
	return def
}

var UserAgentIDCookieDef = NewUserAgentIDCookieDef()

func CollectCookies(ctx context.Context, deps *Dependencies, workflows Workflows) (cookies []*http.Cookie, err error) {
	err = TraverseWorkflow(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			if n, ok := nodeSimple.(CookieGetter); ok {
				c, err := n.GetCookies(ctx, deps, workflows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
		Intent: func(intent Intent, w *Workflow) error {
			if i, ok := intent.(CookieGetter); ok {
				c, err := i.GetCookies(ctx, deps, workflows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
	}, workflows.Nearest)
	if err != nil {
		return
	}

	return
}
