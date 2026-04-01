package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps")
}

type AppsListHandler struct {
	// Add service dependencies here as needed
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

// TODO: Replace dummy data with real implementation.
var dummyApps = []siteadmin.App{
	{
		Id:         "app-alpha",
		OwnerEmail: "alice@example.com",
		Plan:       "enterprise",
		CreatedAt:  time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
	},
	{
		Id:         "app-beta",
		OwnerEmail: "bob@example.com",
		Plan:       "startups",
		CreatedAt:  time.Date(2024, 3, 22, 10, 30, 0, 0, time.UTC),
	},
	{
		Id:         "app-gamma",
		OwnerEmail: "carol@example.com",
		Plan:       "free",
		CreatedAt:  time.Date(2024, 6, 5, 14, 0, 0, 0, time.UTC),
	},
	{
		Id:         "app-delta",
		OwnerEmail: "alice@example.com",
		Plan:       "startups",
		CreatedAt:  time.Date(2024, 8, 10, 9, 0, 0, 0, time.UTC),
	},
	{
		Id:         "app-epsilon",
		OwnerEmail: "eve@example.com",
		Plan:       "free",
		CreatedAt:  time.Date(2024, 11, 1, 12, 0, 0, 0, time.UTC),
	},
}

func (h *AppsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseAppsListParams(r)

	// TODO: Replace with real data source. Filter and paginate dummy data for now.
	filtered := make([]siteadmin.App, 0, len(dummyApps))
	for _, a := range dummyApps {
		if params.AppID != "" && !strings.EqualFold(a.Id, params.AppID) {
			continue
		}
		if params.OwnerEmail != "" && !strings.EqualFold(a.OwnerEmail, params.OwnerEmail) {
			continue
		}
		filtered = append(filtered, a)
	}

	totalCount := len(filtered)
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	response := siteadmin.AppsListResponse{
		Apps:       filtered[start:end],
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
