package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authapi "github.com/authgear/authgear-server/pkg/auth/api"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureLoginLinkV1ApproveRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/login_link/approve")
}

var LoginLinkV1ApproveHandlerLogger = slogutil.NewLogger("handler-login-link-v1-approve")

type LoginLinkV1ApproveHandler struct {
	Controller *authapi.APIController
}

func (h *LoginLinkV1ApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body map[string]any
	err := httputil.ParseJSONBody(r, w, func(reader io.Reader, value any) error {
		return json.NewDecoder(reader).Decode(value)
	}, &body)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	err = h.Controller.ApproveLoginLink(ctx, body)
	if err != nil {
		if !apierrors.IsAPIError(err) {
			LoginLinkV1ApproveHandlerLogger.GetLogger(ctx).WithError(err).Error(ctx, "failed to approve login link")
		}
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	httputil.WriteJSONResponse(ctx, w, &api.Response{Result: struct{}{}})
}
