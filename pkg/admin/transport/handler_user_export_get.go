package transport

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserExportGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/_api/admin/users/export/:id")
}

type UserExportGetHandlerCloudStorage interface {
	PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error)
}

type UserExportGetProducer interface {
	GetTask(ctx context.Context, item *redisqueue.QueueItem) (*redisqueue.Task, error)
}

type UserExportGetHandler struct {
	AppID        config.AppID
	UserExports  UserExportGetProducer
	CloudStorage UserExportGetHandlerCloudStorage
}

func (h *UserExportGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.handle(ctx, w, r)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}
}

func (h *UserExportGetHandler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.CloudStorage == nil {
		return userexport.ErrUserExportDisabled
	}

	taskID := httproute.GetParam(r, "id")
	queueItem := &redisqueue.QueueItem{
		AppID:  string(h.AppID),
		TaskID: taskID,
	}

	task, err := h.UserExports.GetTask(ctx, queueItem)
	if err != nil {
		return err
	}

	response, err := userexport.NewResponseFromTask(task)
	if err != nil {
		return err
	}

	// Get presigned download url when the task completed successfully
	if response.Status == "completed" && response.FailedAt == nil {
		downloadUrl, err := h.CloudStorage.PresignGetObject(ctx, response.DownloadUrl, userexport.PresignGetExpiresForUserExport)
		if err != nil {
			return err
		}
		response.DownloadUrl = downloadUrl.String()
	}

	httputil.WriteJSONResponse(ctx, w, &api.Response{
		Result: response,
	})
	return nil
}
