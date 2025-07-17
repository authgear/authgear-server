package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureUserExportCreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/export")
}

type UserExportCreateHandlerCloudStorage interface {
	PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error)
}

type UserExportCreateProducer interface {
	NewTask(appID string, input json.RawMessage, taskIDPrefix string) *redisqueue.Task
	EnqueueTask(ctx context.Context, task *redisqueue.Task) error
}

type UserExportCreateHandlerUserExportService interface {
	ParseExportRequest(w http.ResponseWriter, r *http.Request) (*userexport.Request, error)
}

const (
	usageLimitUserExport usage.LimitName = "UserExport"
)

type UserExportUsageLimiter interface {
	Reserve(ctx context.Context, name usage.LimitName, config *config.UsageLimitConfig) (*usage.Reservation, error)
}

type UserExportCreateHandler struct {
	AppID                 config.AppID
	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig
	Producer              UserExportCreateProducer
	UsageLimiter          UserExportUsageLimiter
	CloudStorage          UserExportCreateHandlerCloudStorage
	Service               UserExportCreateHandlerUserExportService
}

func (h *UserExportCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.handle(ctx, w, r)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}
}

func (h *UserExportCreateHandler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.CloudStorage == nil {
		return userexport.ErrUserExportDisabled
	}

	request, err := h.Service.ParseExportRequest(w, r)
	if err != nil {
		return err
	}

	rawMessage, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = h.UsageLimiter.Reserve(
		ctx,
		usageLimitUserExport,
		h.AdminAPIFeatureConfig.UserExportUsage,
	)
	if err != nil {
		return err
	}

	task := h.Producer.NewTask(string(h.AppID), rawMessage, "userexport")
	err = h.Producer.EnqueueTask(ctx, task)
	if err != nil {
		return err
	}

	response, err := userexport.NewResponseFromTask(task)
	if err != nil {
		return err
	}

	httputil.WriteJSONResponse(ctx, w, &api.Response{
		Result: response,
	})
	return nil
}
