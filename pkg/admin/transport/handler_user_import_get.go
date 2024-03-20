package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureUserImportGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/_api/admin/users/import/:id")
}

type UserImportGetProducer interface {
	GetTask(ctx context.Context, item *redisqueue.QueueItem) (*redisqueue.Task, error)
}

type UserImportGetHandler struct {
	AppID       config.AppID
	JSON        JSONResponseWriter
	UserImports UserImportGetProducer
}

func (h *UserImportGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserImportGetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	taskID := httproute.GetParam(r, "id")
	queueItem := &redisqueue.QueueItem{
		AppID:  string(h.AppID),
		TaskID: taskID,
	}

	task, err := h.UserImports.GetTask(r.Context(), queueItem)
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
