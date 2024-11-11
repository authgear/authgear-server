package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserImportCreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/import")
}

type UserImportJobEnqueuer interface {
	EnqueueJob(ctx context.Context, request *userimport.Request) (*userimport.Response, error)
}

type UserImportCreateHandler struct {
	JSON        JSONResponseWriter
	UserImports UserImportJobEnqueuer
}

func (h *UserImportCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.handle(ctx, w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserImportCreateHandler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request userimport.Request
	err := httputil.BindJSONBody(r, w, userimport.RequestSchema.Validator(), &request, httputil.WithBodyMaxSize(userimport.BodyMaxSize))
	if err != nil {
		return err
	}

	resp, err := h.UserImports.EnqueueJob(ctx, &request)
	if err != nil {
		return err
	}

	h.JSON.WriteResponse(w, &api.Response{
		Result: resp,
	})
	return nil
}
