package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigurePlansListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/plans")
}

type PlansListService interface {
	ListPlans(ctx context.Context) ([]siteadmin.Plan, error)
}

type PlansListHandler struct {
	Service PlansListService
}

func (h *PlansListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	plans, err := h.Service.ListPlans(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	SiteAdminAPISuccessResponse{Body: siteadmin.PlansListResponse{Plans: plans}}.WriteTo(w)
}
