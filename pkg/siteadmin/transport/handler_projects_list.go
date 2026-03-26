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

func ConfigureProjectsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects")
}

type ProjectsListHandler struct {
	// Add service dependencies here as needed
}

type ProjectsListParams struct {
	Page       int
	PageSize   int
	ProjectID  string
	OwnerEmail string
}

func parseProjectsListParams(r *http.Request) ProjectsListParams {
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

	return ProjectsListParams{
		Page:       page,
		PageSize:   pageSize,
		ProjectID:  q.Get("project_id"),
		OwnerEmail: q.Get("owner_email"),
	}
}

// TODO: Replace dummy data with real implementation.
var dummyProjects = []siteadmin.Project{
	{
		Id:         "project-alpha",
		OwnerEmail: "alice@example.com",
		Plan:       "enterprise",
		CreatedAt:  time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
	},
	{
		Id:         "project-beta",
		OwnerEmail: "bob@example.com",
		Plan:       "startups",
		CreatedAt:  time.Date(2024, 3, 22, 10, 30, 0, 0, time.UTC),
	},
	{
		Id:         "project-gamma",
		OwnerEmail: "carol@example.com",
		Plan:       "free",
		CreatedAt:  time.Date(2024, 6, 5, 14, 0, 0, 0, time.UTC),
	},
	{
		Id:         "project-delta",
		OwnerEmail: "alice@example.com",
		Plan:       "startups",
		CreatedAt:  time.Date(2024, 8, 10, 9, 0, 0, 0, time.UTC),
	},
	{
		Id:         "project-epsilon",
		OwnerEmail: "eve@example.com",
		Plan:       "free",
		CreatedAt:  time.Date(2024, 11, 1, 12, 0, 0, 0, time.UTC),
	},
}

func (h *ProjectsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseProjectsListParams(r)

	// TODO: Replace with real data source. Filter and paginate dummy data for now.
	filtered := make([]siteadmin.Project, 0, len(dummyProjects))
	for _, p := range dummyProjects {
		if params.ProjectID != "" && !strings.EqualFold(p.Id, params.ProjectID) {
			continue
		}
		if params.OwnerEmail != "" && !strings.EqualFold(p.OwnerEmail, params.OwnerEmail) {
			continue
		}
		filtered = append(filtered, p)
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

	response := siteadmin.ProjectsListResponse{
		Projects:   filtered[start:end],
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
