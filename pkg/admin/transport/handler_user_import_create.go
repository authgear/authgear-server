package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserImportCreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/import")
}

type UserImportCreateProducer interface {
	NewTask(appID string, input json.RawMessage, taskIDPrefix string) *redisqueue.Task
	EnqueueTask(ctx context.Context, task *redisqueue.Task) error
}

type UserImportCreateHandler struct {
	AppID       config.AppID
	JSON        JSONResponseWriter
	UserImports UserImportCreateProducer
}

func (h *UserImportCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserImportCreateHandler) handle(w http.ResponseWriter, r *http.Request) error {
	var request userimport.Request
	err := httputil.BindJSONBody(r, w, userimport.RequestSchema.Validator(), &request, httputil.WithBodyMaxSize(userimport.BodyMaxSize))
	if err != nil {
		return err
	}

	rawMessage, err := json.Marshal(request)
	if err != nil {
		return err
	}

	task := h.UserImports.NewTask(string(h.AppID), rawMessage, "task")
	err = h.UserImports.EnqueueTask(r.Context(), task)
	if err != nil {
		return err
	}

	response, err := userimport.NewResponseFromTask(task)
	if err != nil {
		return err
	}

	h.JSON.WriteResponse(w, &api.Response{
		Result: response,
	})
	return nil
}
