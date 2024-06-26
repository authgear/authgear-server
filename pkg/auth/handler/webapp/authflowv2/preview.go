package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowV2PreviewRoute(route httproute.Route) httproute.Route {
	return route.PrependPathPattern(webapp.InlinePreviewPathPrefix).WithMethods(http.MethodGet)
}
