package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

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

type ProjectsListResponse struct {
	Projects []interface{} `json:"projects"`
}

func (h *ProjectsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = parseProjectsListParams(r)

	// TODO: Replace with real data source.
	response := ProjectsListResponse{
		Projects: []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
