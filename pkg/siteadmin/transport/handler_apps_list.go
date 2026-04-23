package transport

import (
	"context"
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
	Page       uint64
	PageSize   uint64
	AppID      string
	OwnerEmail string
	Plan       string
	Sort       siteadmin.ListAppsParamsSort
	Order      siteadmin.ListAppsParamsOrder
}

func parseAppsListParams(r *http.Request) AppsListParams {
	q := r.URL.Query()

	page := uint64(1)
	if v := q.Get("page"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n >= 1 {
			page = n
		}
	}

	pageSize := uint64(service.MaxPageSize)
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n >= 1 {
			if n > service.MaxPageSize {
				pageSize = service.MaxPageSize
			} else {
				pageSize = n
			}
		}
	}

	sortVal := siteadmin.ListAppsParamsSort(q.Get("sort"))
	if !sortVal.Valid() {
		sortVal = siteadmin.CreatedAt
	}

	orderVal := siteadmin.ListAppsParamsOrder(q.Get("order"))
	if !orderVal.Valid() {
		orderVal = siteadmin.Desc
	}

	return AppsListParams{
		Page:       page,
		PageSize:   pageSize,
		AppID:      q.Get("app_id"),
		OwnerEmail: q.Get("owner_email"),
		Plan:       q.Get("plan"),
		Sort:       sortVal,
		Order:      orderVal,
	}
}

func (h *AppsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseAppsListParams(r)

	result, err := h.AppsList.ListApps(r.Context(), service.ListAppsParams{
		Page:       params.Page,
		PageSize:   params.PageSize,
		AppID:      params.AppID,
		OwnerEmail: params.OwnerEmail,
		Plan:       params.Plan,
		Sort:       params.Sort,
		Order:      params.Order,
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
	SiteAdminAPISuccessResponse{Body: response}.WriteTo(w)
}
