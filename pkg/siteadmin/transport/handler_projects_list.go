package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureProjectsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects")
}

type ProjectsListHandler struct {
	// Add service dependencies here as needed
}

type ProjectsListResponse struct {
	Projects []interface{} `json:"projects"`
}

func (h *ProjectsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Scaffolding: return empty list for now
	response := ProjectsListResponse{
		Projects: []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
