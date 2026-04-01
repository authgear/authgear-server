package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	service "github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps")
}

type AppsListService interface {
	ListApps(ctx context.Context, params service.ListAppsParams) (*service.ListAppsResult, error)
}

type AppsListHandler struct {
	AppsList AppsListService
}

type AppsListParams struct {
	Page       int
	PageSize   int
	AppID      string
	OwnerEmail string
}

func parseAppsListParams(r *http.Request) AppsListParams {
	q := r.URL.Query()

	page := 1
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}

	pageSize := 20
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
			pageSize = n
		}
	}

	return AppsListParams{
		Page:       page,
		PageSize:   pageSize,
		AppID:      q.Get("app_id"),
		OwnerEmail: q.Get("owner_email"),
	}
}

func (h *AppsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseAppsListParams(r)

	result, err := h.AppsList.ListApps(r.Context(), service.ListAppsParams{
		Page:       params.Page,
		PageSize:   params.PageSize,
		AppID:      params.AppID,
		OwnerEmail: params.OwnerEmail,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := siteadmin.AppsListResponse{
		Apps:       result.Apps,
		TotalCount: result.TotalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
