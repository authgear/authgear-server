package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/admin/facade"
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserExportCreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/export")
}

type UserExportCreateHandler struct {
	JSON JSONResponseWriter
	// TODO: Replace facade by worker task
	User facade.UserExportFacade
}

func (h *UserExportCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserExportCreateHandler) handle(w http.ResponseWriter, r *http.Request) error {
	var request userexport.Request
	err := httputil.BindJSONBody(r, w, userexport.RequestSchema.Validator(), &request)
	if err != nil {
		return err
	}

	// TODO: change direct call to be a worker task
	response := h.User.ExportRecords(nil, nil)

	h.JSON.WriteResponse(w, &api.Response{
		Result: response,
	})
	return nil
}
