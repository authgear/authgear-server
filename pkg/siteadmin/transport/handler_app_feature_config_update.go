package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAppFeatureConfigUpdateRoute(route httproute.Route) httproute.Route {
	// The OPTIONS request for this shared path is handled by
	// ConfigureAppFeatureConfigGetRoute.
	return route.WithMethods("PUT").
		WithPathPattern("/api/v1/apps/:appID/feature-config")
}

var AppFeatureConfigUpdateRequestSchema = validation.NewSimpleSchema(`
{
    "type": "object",
    "properties": { "app_feature_config_yaml": { "type": "string" } },
    "required": ["app_feature_config_yaml"]
}
`)

type AppFeatureConfigUpdateService interface {
	UpdateAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*service.AppFeatureConfigResult, error)
}

type AppFeatureConfigUpdateHandler struct {
	Service AppFeatureConfigUpdateService
}

type AppFeatureConfigUpdateParams struct {
	AppID string
	siteadmin.UpdateAppFeatureConfigRequest
}

func parseAppFeatureConfigUpdateParams(r *http.Request) (AppFeatureConfigUpdateParams, error) {
	var body siteadmin.UpdateAppFeatureConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return AppFeatureConfigUpdateParams{}, err
	}
	if err := AppFeatureConfigUpdateRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return AppFeatureConfigUpdateParams{}, err
	}
	return AppFeatureConfigUpdateParams{
		AppID:                         httproute.GetParam(r, "appID"),
		UpdateAppFeatureConfigRequest: body,
	}, nil
}

func (h *AppFeatureConfigUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseAppFeatureConfigUpdateParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	result, err := h.Service.UpdateAppFeatureConfig(r.Context(), params.AppID, params.AppFeatureConfigYaml)
	if err != nil {
		writeError(w, r, err)
		return
	}
	SiteAdminAPISuccessResponse{Body: featureConfigResultToResponse(result)}.WriteTo(w)
}
