package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppFeatureConfigGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps/:appID/feature-config")
}

type AppFeatureConfigGetService interface {
	GetAppFeatureConfig(ctx context.Context, appID string) (*service.AppFeatureConfigResult, error)
}

type AppFeatureConfigGetHandler struct {
	Service AppFeatureConfigGetService
}

func (h *AppFeatureConfigGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	appID := httproute.GetParam(r, "appID")

	result, err := h.Service.GetAppFeatureConfig(r.Context(), appID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: featureConfigResultToResponse(result)}.WriteTo(w)
}

// featureConfigResultToResponse is shared by the get/update/preview handlers
// (defined here since this is the first of the three files registered).
func featureConfigResultToResponse(result *service.AppFeatureConfigResult) siteadmin.AppFeatureConfigResponse {
	return siteadmin.AppFeatureConfigResponse{
		PlanName:                   result.PlanName,
		EffectivePlanFeatureConfig: *result.EffectivePlanFeatureConfig,
		AppFeatureConfigYaml:       result.AppFeatureConfigYAML,
		EffectiveAppFeatureConfig:  *result.EffectiveAppFeatureConfig,
	}
}
