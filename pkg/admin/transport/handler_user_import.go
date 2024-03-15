package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserImportRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/import")
}

type UserImportService interface {
	ImportRecords(ctx context.Context, request *userimport.Request) (*userimport.Summary, []userimport.Detail)
}

type UserImportHandler struct {
	JSON              JSONResponseWriter
	UserImportService UserImportService
}

func (h *UserImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request userimport.Request
	err = httputil.BindJSONBody(r, w, userimport.RequestSchema.Validator(), &request, httputil.WithBodyMaxSize(userimport.BodyMaxSize))
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	h.handle(w, r, &request)
}

func (h *UserImportHandler) handle(w http.ResponseWriter, r *http.Request, request *userimport.Request) {
	summary, details := h.UserImportService.ImportRecords(r.Context(), request)
	result := map[string]interface{}{
		"summary": summary,
	}
	if len(details) > 0 {
		result["details"] = details
	}
	h.JSON.WriteResponse(w, &api.Response{
		Result: result,
	})
}
